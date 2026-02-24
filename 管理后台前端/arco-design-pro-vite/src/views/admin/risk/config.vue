<template>
  <div class="risk-config-container">
    <a-card class="config-card" :bordered="false">
      <template #title>
        <div class="header">
          <span>全局风控参数设置</span>
          <a-button :loading="loading" @click="handleRefresh">
            <template #icon><icon-sync /></template>
            刷新
          </a-button>
        </div>
      </template>

      <a-spin :loading="loading">
        <a-form
          :model="form"
          class="form-content"
          layout="vertical"
          @submit="handleSubmit"
        >
          <!-- 默认胜率配置 移动到最上方 -->
          <div class="section">
            <h3 class="section-title">默认胜率配置</h3>
            <p class="section-desc">
              当未匹配到特定的金额规则或群控预设时，系统将采用以下默认概率执行。
            </p>
            <div class="probability-grid">
              <a-form-item
                label="默认盈利概率 (%)"
                help="未匹配到规则时的通用盈利胜率"
              >
                <a-input-number
                  v-model="form.risk_profit_probability"
                  :min="0"
                  :max="100"
                  @blur="
                    updateSetting(
                      'risk_profit_probability',
                      form.risk_profit_probability
                    )
                  "
                />
              </a-form-item>
              <a-form-item
                label="默认亏损概率 (%)"
                help="自动计算：100% - 盈利概率"
              >
                <a-input-number
                  :model-value="100 - form.risk_profit_probability"
                  disabled
                />
              </a-form-item>
            </div>
            <div class="fixed-info">
              <a-tag color="arcoblue">
                <template #icon><icon-info-circle-fill /></template>
                提前风控秒数已固定为 5 秒
              </a-tag>
            </div>
          </div>

          <a-divider />

          <a-form-item label="当前风控模式" help="选择系统全局采用的风控策略">
            <a-radio-group
              v-model="form.risk_mode"
              type="button"
              @change="(val) => updateSetting('risk_mode', val)"
            >
              <a-radio value="0">无风控</a-radio>
              <a-radio value="2">群控 (全局结果控制)</a-radio>
              <a-radio value="3">金额控制 (按订单金额概率控制)</a-radio>
            </a-radio-group>
          </a-form-item>

          <a-divider />

          <!-- 群控配置 -->
          <div v-if="form.risk_mode === '2'" class="section">
            <h3 class="section-title">群控结果配置</h3>
            <a-form-item
              label="群控预设结果"
              help="开启群控模式后，所有订单将默认按此结果执行"
            >
              <a-select
                v-model="form.risk_group_result"
                style="width: 200px"
                @change="(val) => updateSetting('risk_group_result', val)"
              >
                <a-option value="1">全员盈利</a-option>
                <a-option value="-1">全员亏损</a-option>
              </a-select>
            </a-form-item>
          </div>

          <!-- 金额控制配置 -->
          <div v-if="form.risk_mode === '3'" class="section">
            <h3 class="section-title">金额阶梯胜率控制</h3>
            <p class="section-desc">
              根据订单价值（USDT）自动匹配对应的盈利概率。适用于：永续合约、交割合约（秒合约）。
            </p>
            <a-table :data="moneyRules" :pagination="false" :bordered="false">
              <template #columns>
                <a-table-column title="最小金额 (USDT)" data-index="min">
                  <template #cell="{ record }">
                    <a-input-number
                      v-model="record.min"
                      :min="0"
                      size="small"
                    />
                  </template>
                </a-table-column>
                <a-table-column title="最大金额 (USDT)" data-index="max">
                  <template #cell="{ record }">
                    <a-input-number
                      v-model="record.max"
                      :min="0"
                      size="small"
                      placeholder="0为无上限"
                    />
                  </template>
                </a-table-column>
                <a-table-column title="盈利概率 (%)" data-index="probability">
                  <template #cell="{ record }">
                    <a-input-number
                      v-model="record.probability"
                      :min="0"
                      :max="100"
                      size="small"
                    />
                  </template>
                </a-table-column>
                <a-table-column title="操作" align="center">
                  <template #cell="{ index }">
                    <a-button
                      type="text"
                      status="danger"
                      @click="removeMoneyRule(index)"
                    >
                      <template #icon><icon-delete /></template>
                    </a-button>
                  </template>
                </a-table-column>
              </template>
            </a-table>
            <div class="table-actions">
              <a-button type="outline" size="small" long @click="addMoneyRule">
                <template #icon><icon-plus /></template>
                添加金额区间规则
              </a-button>
            </div>
            <div class="save-actions">
              <a-button
                type="primary"
                :loading="savingMoneyRules"
                @click="saveMoneyRules"
              >
                保存金额规则
              </a-button>
            </div>
          </div>
        </a-form>
      </a-spin>
    </a-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    IconSync,
    IconInfoCircleFill,
    IconPlus,
    IconDelete,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';

  interface MoneyRule {
    min: number;
    max: number;
    probability: number;
  }

  const loading = ref(false);
  const savingMoneyRules = ref(false);
  const moneyRules = ref<MoneyRule[]>([]);
  const form = reactive({
    risk_mode: '0',
    risk_group_result: '1',
    risk_money_profit_probability: '',
    risk_profit_probability: 50,
  });

  const parseMoneyRules = (str: string) => {
    if (!str) return [];
    try {
      return str.split('|').map((item) => {
        const [range, prob] = item.split(':');
        const [min, max] = range.split('-');
        return {
          min: Number(min) || 0,
          max: Number(max) || 0,
          probability: Number(prob) || 0,
        };
      });
    } catch (e) {
      console.error('Parse money rules failed:', e);
      return [];
    }
  };

  const serializeMoneyRules = (rules: MoneyRule[]) => {
    return rules.map((r) => `${r.min}-${r.max}:${r.probability}`).join('|');
  };

  const fetchConfig = async () => {
    try {
      loading.value = true;
      const response = await axios.get('/admin/risk/config');
      if (response.data && response.data.data) {
        const { data } = response.data;
        data.forEach((item: any) => {
          if (Object.prototype.hasOwnProperty.call(form, item.key)) {
            // 对数值型字段进行转换
            if (['risk_profit_probability'].includes(item.key)) {
              (form as any)[item.key] = Number(item.value) || 0;
            } else {
              (form as any)[item.key] = item.value;
            }
          }
        });
        // 特殊处理金额规则
        moneyRules.value = parseMoneyRules(form.risk_money_profit_probability);
      }
    } catch (error: any) {
      Message.error(error.message || '获取配置失败');
    } finally {
      loading.value = false;
    }
  };

  const updateSetting = async (key: string, value: any) => {
    try {
      await axios.post('/admin/risk/config', {
        key,
        value: String(value),
      });
      Message.success('配置已更新');
    } catch (error: any) {
      Message.error(error.message || '更新失败');
    }
  };

  const addMoneyRule = () => {
    moneyRules.value.push({ min: 0, max: 0, probability: 0 });
  };

  const removeMoneyRule = (index: number) => {
    moneyRules.value.splice(index, 1);
  };

  const saveMoneyRules = async () => {
    const serialized = serializeMoneyRules(moneyRules.value);
    savingMoneyRules.value = true;
    try {
      await updateSetting('risk_money_profit_probability', serialized);
      form.risk_money_profit_probability = serialized;
    } finally {
      savingMoneyRules.value = false;
    }
  };

  const handleRefresh = () => {
    fetchConfig();
  };

  const handleSubmit = () => {
    Message.success('设置已保存');
  };

  fetchConfig();
</script>

<style scoped lang="less">
  .risk-config-container {
    padding: 20px;
    background-color: var(--color-fill-2);
    min-height: 100%;

    .config-card {
      border-radius: 8px;

      .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }

      .form-content {
        max-width: 900px;
        margin: 0 auto;
        padding: 20px 0;
      }

      .section {
        margin-bottom: 24px;

        .section-title {
          margin-bottom: 8px;
          font-size: 16px;
          color: var(--color-text-1);
          border-left: 4px solid var(--color-primary-light-4);
          padding-left: 12px;
        }

        .section-desc {
          font-size: 13px;
          color: var(--color-text-3);
          margin-bottom: 16px;
          padding-left: 16px;
        }
      }

      .table-actions {
        margin-top: 12px;
      }

      .save-actions {
        margin-top: 20px;
        text-align: right;
      }

      .probability-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
      }

      .fixed-info {
        margin-top: 10px;
      }
    }
  }
</style>
