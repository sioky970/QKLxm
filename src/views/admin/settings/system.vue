<template>
  <div class="system-settings-container">
    <a-tabs v-model:active-key="activeTab" type="card">
      <!-- 系统设置 Tab -->
      <a-tab-pane key="system" title="系统设置">
        <a-card class="settings-card" :bordered="false">
          <template #title>
            <div class="table-header">
              <span>系统业务设置</span>
            </div>
          </template>

          <a-spin :loading="loading" style="width: 100%">
            <a-form :model="settings" layout="vertical" @submit="handleSave">
              <a-divider orientation="left">注册设置</a-divider>

              <a-form-item
                label="注册必须使用邀请码"
                help="开启后，用户注册时必须填写有效的邀请码。关闭后，邀请码为选填项。"
              >
                <a-switch
                  v-model="settings.register_mandatory_invitation"
                  :checked-value="'1'"
                  :unchecked-value="'0'"
                  type="round"
                >
                  <template #checked>必填</template>
                  <template #unchecked>选填</template>
                </a-switch>
              </a-form-item>

              <a-divider />

              <a-form-item>
                <a-space>
                  <a-button
                    type="primary"
                    html-type="submit"
                    :loading="saveLoading"
                  >
                    <template #icon><icon-save /></template>
                    保存设置
                  </a-button>
                  <a-button @click="fetchSettings">
                    <template #icon><icon-refresh /></template>
                    重置
                  </a-button>
                </a-space>
              </a-form-item>
            </a-form>
          </a-spin>
        </a-card>
      </a-tab-pane>

      <!-- 提现设置 Tab -->
      <a-tab-pane key="withdraw" title="提现设置">
        <a-card class="settings-card" :bordered="false">
          <template #title>
            <div class="table-header">
              <span>提现模式配置</span>
            </div>
          </template>

          <a-spin :loading="withdrawLoading" style="width: 100%">
            <a-form :model="withdrawSettings" layout="vertical" @submit="handleSaveWithdraw">
              <a-divider orientation="left">提现模式</a-divider>

              <a-form-item
                label="提现模式"
                help="设置用户可使用的提现方式。银行转账需用户绑定银行卡，区块链提币需用户输入钱包地址。"
              >
                <a-radio-group v-model="withdrawSettings.withdraw_mode" direction="vertical">
                  <a-radio value="bank">
                    <div class="radio-label">
                      <icon-idcard class="radio-icon" />
                      <span>仅银行转账</span>
                    </div>
                    <div class="radio-desc">用户只能提现到银行卡</div>
                  </a-radio>
                  <a-radio value="crypto">
                    <div class="radio-label">
                      <icon-send class="radio-icon" />
                      <span>仅区块链提币</span>
                    </div>
                    <div class="radio-desc">用户只能提现到区块链钱包地址</div>
                  </a-radio>
                  <a-radio value="both">
                    <div class="radio-label">
                      <icon-apps class="radio-icon" />
                      <span>两种都支持</span>
                    </div>
                    <div class="radio-desc">用户可自由选择银行卡或区块链提币</div>
                  </a-radio>
                </a-radio-group>
              </a-form-item>

              <a-divider orientation="left">区块链网络配置</a-divider>

              <a-form-item
                label="支持的区块链网络"
                help="选择允许用户提币的区块链网络类型"
              >
                <a-checkbox-group v-model="withdrawSettings.crypto_networks_list">
                  <a-checkbox value="TRC20">
                    <span class="network-label">TRC20</span>
                    <span class="network-desc">Tron网络，手续费低</span>
                  </a-checkbox>
                  <a-checkbox value="ERC20">
                    <span class="network-label">ERC20</span>
                    <span class="network-desc">以太坊网络，广泛支持</span>
                  </a-checkbox>
                  <a-checkbox value="BEP20">
                    <span class="network-label">BEP20</span>
                    <span class="network-desc">币安智能链网络</span>
                  </a-checkbox>
                  <a-checkbox value="Polygon">
                    <span class="network-label">Polygon</span>
                    <span class="network-desc">Polygon/Matic网络</span>
                  </a-checkbox>
                  <a-checkbox value="BRC20">
                    <span class="network-label">BRC20</span>
                    <span class="network-desc">比特币网络</span>
                  </a-checkbox>
                </a-checkbox-group>
              </a-form-item>

              <a-form-item
                label="区块链提币手续费(USDT)"
                help="每笔区块链提币收取的手续费"
              >
                <a-input-number
                  v-model="withdrawSettings.crypto_withdraw_fee"
                  :min="0"
                  :precision="2"
                  :step="0.1"
                  style="width: 200px"
                >
                  <template #suffix>USDT</template>
                </a-input-number>
              </a-form-item>

              <a-divider />

              <a-form-item>
                <a-space>
                  <a-button
                    type="primary"
                    html-type="submit"
                    :loading="withdrawSaveLoading"
                  >
                    <template #icon><icon-save /></template>
                    保存设置
                  </a-button>
                  <a-button @click="fetchWithdrawSettings">
                    <template #icon><icon-refresh /></template>
                    重置
                  </a-button>
                </a-space>
              </a-form-item>
            </a-form>
          </a-spin>
        </a-card>
      </a-tab-pane>

      <!-- 充值地址管理 Tab -->
      <a-tab-pane key="deposit" title="充值地址管理">
        <DepositAddressManager />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { IconSave, IconRefresh, IconIdcard, IconSend, IconApps } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import DepositAddressManager from './deposit-address.vue';

  const activeTab = ref('system');
  const loading = ref(false);
  const saveLoading = ref(false);

  // 提现设置相关
  const withdrawLoading = ref(false);
  const withdrawSaveLoading = ref(false);

  const settings = reactive({
    register_mandatory_invitation: '0',
  });

  const withdrawSettings = reactive({
    withdraw_mode: 'both',
    crypto_networks_list: ['TRC20', 'ERC20', 'BEP20', 'Polygon', 'BRC20'] as string[],
    crypto_withdraw_fee: 1,
  });

  const fetchSettings = async () => {
    try {
      loading.value = true;
      const response = await axios.post('/admin/config/list');
      if (response.data && response.data.data) {
        const list = response.data.data || [];
        const mandatorySetting = list.find(
          (item: any) => item.key === 'register_mandatory_invitation'
        );
        if (mandatorySetting) {
          settings.register_mandatory_invitation = mandatorySetting.value;
        } else {
          settings.register_mandatory_invitation = '0';
        }
      }
    } catch (error: any) {
      Message.error(error.message || '获取设置失败');
    } finally {
      loading.value = false;
    }
  };

  const fetchWithdrawSettings = async () => {
    try {
      withdrawLoading.value = true;
      const response = await axios.post('/admin/config/list');
      if (response.data && response.data.data) {
        const list = response.data.data || [];
        
        // 提现模式
        const modeSetting = list.find((item: any) => item.key === 'withdraw_mode');
        if (modeSetting) {
          withdrawSettings.withdraw_mode = modeSetting.value || 'both';
        }
        
        // 区块链网络
        const networksSetting = list.find((item: any) => item.key === 'crypto_networks');
        if (networksSetting && networksSetting.value) {
          withdrawSettings.crypto_networks_list = networksSetting.value.split(',').filter((n: string) => n);
        }
        
        // 手续费
        const feeSetting = list.find((item: any) => item.key === 'crypto_withdraw_fee');
        if (feeSetting) {
          withdrawSettings.crypto_withdraw_fee = parseFloat(feeSetting.value) || 1;
        }
      }
    } catch (error: any) {
      Message.error(error.message || '获取提现设置失败');
    } finally {
      withdrawLoading.value = false;
    }
  };

  const handleSave = async () => {
    try {
      saveLoading.value = true;
      await axios.post('/admin/config/update', {
        key: 'register_mandatory_invitation',
        value: settings.register_mandatory_invitation,
      });
      Message.success('保存成功');
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    } finally {
      saveLoading.value = false;
    }
  };

  const handleSaveWithdraw = async () => {
    try {
      withdrawSaveLoading.value = true;
      
      // 保存提现模式
      await axios.post('/admin/config/update', {
        key: 'withdraw_mode',
        value: withdrawSettings.withdraw_mode,
      });
      
      // 保存区块链网络
      await axios.post('/admin/config/update', {
        key: 'crypto_networks',
        value: withdrawSettings.crypto_networks_list.join(','),
      });
      
      // 保存手续费
      await axios.post('/admin/config/update', {
        key: 'crypto_withdraw_fee',
        value: String(withdrawSettings.crypto_withdraw_fee),
      });
      
      Message.success('提现设置保存成功');
    } catch (error: any) {
      Message.error(error.message || '保存失败');
    } finally {
      withdrawSaveLoading.value = false;
    }
  };

  onMounted(() => {
    fetchSettings();
    fetchWithdrawSettings();
  });
</script>

<style scoped lang="less">
  .system-settings-container {
    padding: 20px;

    :deep(.arco-tabs-content) {
      padding-top: 16px;
    }

    .settings-card {
      :deep(.arco-card-body) {
        padding: 24px;
        max-width: 800px;
      }

      .table-header {
        font-size: 16px;
        font-weight: 500;
      }
    }

    .radio-label {
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 500;
      
      .radio-icon {
        font-size: 16px;
        color: #165dff;
      }
    }

    .radio-desc {
      font-size: 12px;
      color: #86909c;
      margin-left: 24px;
      margin-top: 4px;
    }

    .network-label {
      font-weight: 500;
      margin-right: 8px;
    }

    .network-desc {
      font-size: 12px;
      color: #86909c;
    }

    :deep(.arco-checkbox) {
      margin-bottom: 12px;
    }

    :deep(.arco-radio) {
      margin-bottom: 16px;
    }
  }
</style>
