<template>
  <div class="user-list-container">
    <!-- 搜索筛选区 -->
    <a-card class="search-card" :bordered="false">
      <a-row :gutter="16">
        <a-col :flex="1">
          <a-form :model="searchForm" layout="inline" class="search-form">
            <a-form-item field="keyword" hide-label>
              <a-input
                v-model="searchForm.keyword"
                placeholder="搜索账号/手机/邮箱"
                allow-clear
                style="width: 200px"
              >
                <template #prefix><icon-search /></template>
              </a-input>
            </a-form-item>
            <a-form-item field="status" hide-label>
              <a-select
                v-model="searchForm.status"
                placeholder="用户状态"
                style="width: 120px"
                allow-clear
              >
                <a-option :value="0">正常</a-option>
                <a-option :value="1">已锁定</a-option>
              </a-select>
            </a-form-item>
            <a-form-item field="is_realname" hide-label>
              <a-select
                v-model="searchForm.is_realname"
                placeholder="认证状态"
                style="width: 120px"
                allow-clear
              >
                <a-option :value="1">未认证</a-option>
                <a-option :value="2">已认证</a-option>
              </a-select>
            </a-form-item>
            <a-form-item field="risk" hide-label>
              <a-select
                v-model="searchForm.risk"
                placeholder="风控类型"
                style="width: 120px"
                allow-clear
              >
                <a-option :value="0">无风控</a-option>
                <a-option :value="-1">亏损</a-option>
                <a-option :value="1">盈利</a-option>
              </a-select>
            </a-form-item>
            <!-- 高级筛选（可折叠） -->
            <template v-if="showAdvanced">
              <a-form-item field="real_name" hide-label>
                <a-input
                  v-model="searchForm.real_name"
                  placeholder="真实姓名"
                  allow-clear
                  style="width: 140px"
                />
              </a-form-item>
            </template>
          </a-form>
        </a-col>
        <a-col :flex="'280px'" class="search-actions">
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <template #icon><icon-search /></template>
              搜索
            </a-button>
            <a-button @click="handleReset">
              <template #icon><icon-refresh /></template>
              重置
            </a-button>
            <a-button type="text" @click="showAdvanced = !showAdvanced">
              {{ showAdvanced ? '收起' : '展开' }}
              <icon-up v-if="showAdvanced" />
              <icon-down v-else />
            </a-button>
          </a-space>
        </a-col>
      </a-row>
    </a-card>

    <!-- 表格区 -->
    <a-card class="table-card" :bordered="false">
      <!-- 工具栏 -->
      <div class="table-toolbar">
        <div class="toolbar-left">
          <a-space>
            <span class="table-title">用户列表</span>
            <a-tag color="arcoblue">{{ pagination.total }} 条记录</a-tag>
          </a-space>
        </div>
        <div class="toolbar-right">
          <a-space>
            <!-- 批量操作按钮（选中时显示） -->
            <template v-if="selectedRowKeys.length > 0">
              <a-tag color="blue" size="large">
                已选 {{ selectedRowKeys.length }} 项
              </a-tag>
              <a-button
                type="outline"
                status="warning"
                size="small"
                @click="handleBatchRisk"
              >
                <template #icon><icon-thunderbolt /></template>
                批量风控
              </a-button>
              <a-button
                type="outline"
                size="small"
                @click="handleBatchStatus(0)"
              >
                <template #icon><icon-unlock /></template>
                批量解锁
              </a-button>
              <a-button
                type="outline"
                status="danger"
                size="small"
                @click="handleBatchStatus(1)"
              >
                <template #icon><icon-lock /></template>
                批量锁定
              </a-button>
              <a-divider direction="vertical" />
            </template>
            <a-button type="primary" @click="handleRefresh">
              <template #icon><icon-sync /></template>
              刷新
            </a-button>
            <a-dropdown>
              <a-button>
                <template #icon><icon-more-vertical /></template>
              </a-button>
              <template #content>
                <a-doption @click="handleExport">
                  <template #icon><icon-download /></template>
                  导出数据
                </a-doption>
                <a-doption @click="handleColumnSetting">
                  <template #icon><icon-settings /></template>
                  列设置
                </a-doption>
              </template>
            </a-dropdown>
          </a-space>
        </div>
      </div>

      <!-- 数据表格 -->
      <a-table
        :columns="visibleColumns"
        :data="tableData"
        :pagination="pagination"
        :loading="loading"
        :row-selection="rowSelection"
        :scroll="{ x: tableScrollWidth }"
        :bordered="{ cell: true }"
        stripe
        row-key="id"
        class="user-table"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <!-- ID列 -->
        <template #id="{ record }">
          <span class="cell-id">{{ record.id }}</span>
        </template>

        <!-- 用户信息列（合并账号+认证状态） -->
        <template #userInfo="{ record }">
          <div class="user-info-cell">
            <div class="user-main">
              <span class="account">{{ record.account_number }}</span>
              <a-tag
                v-if="isRealVerified(record)"
                color="green"
                size="small"
                class="verify-tag"
              >
                <icon-check-circle /> 已认证
              </a-tag>
              <a-tag v-else color="gray" size="small" class="verify-tag"
                >未认证</a-tag
              >
            </div>
            <div class="user-sub">
              <span v-if="record.nickname" class="nickname">{{
                record.nickname
              }}</span>
              <span v-if="record.real_name" class="realname">{{
                record.real_name
              }}</span>
            </div>
          </div>
        </template>

        <!-- 联系方式列（合并手机+邮箱） -->
        <template #contact="{ record }">
          <div class="contact-cell">
            <div v-if="record.phone" class="contact-item">
              <icon-phone class="contact-icon" />
              <span>{{ record.phone }}</span>
            </div>
            <div v-if="record.email" class="contact-item">
              <icon-email class="contact-icon" />
              <a-tooltip :content="record.email">
                <span class="email-text">{{ record.email }}</span>
              </a-tooltip>
            </div>
            <span v-if="!record.phone && !record.email" class="empty-text"
              >-</span
            >
          </div>
        </template>

        <!-- 现货USDT余额列 -->
        <template #spotBalance="{ record }">
          <div
            class="balance-cell"
            style="cursor: pointer"
            @click="handleQuickAdjustBalance(record, 'spot')"
          >
            <a-tooltip content="点击调整现货余额">
              <div class="balance-amount">
                <span class="balance-value">{{
                  (record.spot_balance || 0).toFixed(4)
                }}</span>
                <span class="balance-unit">USDT</span>
              </div>
            </a-tooltip>
          </div>
        </template>

        <!-- 合约USDT余额列 -->
        <template #contractBalance="{ record }">
          <div
            class="balance-cell"
            style="cursor: pointer"
            @click="handleQuickAdjustBalance(record, 'contract')"
          >
            <a-tooltip content="点击调整永续合约余额">
              <div class="balance-amount">
                <span class="balance-value">{{
                  (record.contract_balance || 0).toFixed(4)
                }}</span>
                <span class="balance-unit">USDT</span>
              </div>
            </a-tooltip>
          </div>
        </template>

        <!-- 交割合约USDT余额列 -->
        <template #deliveryBalance="{ record }">
          <div
            class="balance-cell"
            style="cursor: pointer"
            @click="handleQuickAdjustBalance(record, 'delivery')"
          >
            <a-tooltip content="点击调整交割合约余额">
              <div class="balance-amount">
                <span class="balance-value">{{
                  (record.delivery_balance || 0).toFixed(4)
                }}</span>
                <span class="balance-unit">USDT</span>
              </div>
            </a-tooltip>
          </div>
        </template>

        <!-- 资金钱包USDT余额列 -->
        <template #fundBalance="{ record }">
          <div
            class="balance-cell"
            style="cursor: pointer"
            @click="handleQuickAdjustBalance(record, 'fund')"
          >
            <a-tooltip content="点击调整资金钱包余额">
              <div class="balance-amount">
                <span class="balance-value fund-balance">{{
                  (record.fund_balance || 0).toFixed(4)
                }}</span>
                <span class="balance-unit">USDT</span>
              </div>
            </a-tooltip>
          </div>
        </template>

        <!-- 状态列 -->
        <template #status="{ record }">
          <a-switch
            :model-value="record.status === 0"
            type="round"
            checked-color="rgb(var(--green-6))"
            unchecked-color="rgb(var(--red-6))"
            @change="(val) => handleQuickToggleStatus(record, val as boolean)"
          >
            <template #checked>正常</template>
            <template #unchecked>锁定</template>
          </a-switch>
        </template>

        <!-- 风控列 -->
        <template #risk="{ record }">
          <a-dropdown
            trigger="click"
            @select="(val: any) => handleQuickSetRisk(record, val)"
          >
            <a-tag
              :color="getRiskColor(record.risk)"
              class="risk-tag"
              style="cursor: pointer"
            >
              <template #icon>
                <icon-thunderbolt v-if="record.risk !== 0" />
              </template>
              {{ getRiskText(record.risk) }}
              <icon-down style="margin-left: 4px; font-size: 10px" />
            </a-tag>
            <template #content>
              <a-doption :value="0">无</a-doption>
              <a-doption :value="1">盈利</a-doption>
              <a-doption :value="-1">亏损</a-doption>
            </template>
          </a-dropdown>
        </template>

        <!-- 邀请关系列 -->
        <template #invite="{ record }">
          <div class="invite-cell">
            <div class="invite-code">
              <span class="label">邀请码:</span>
              <a-tag size="small">{{ record.extension_code || '-' }}</a-tag>
            </div>
            <div v-if="record.parent_id" class="parent-info">
              <span class="label">上级:</span>
              <a-link @click="handleViewParent(record)">
                {{ record.parent_account || `ID:${record.parent_id}` }}
              </a-link>
            </div>
          </div>
        </template>

        <!-- 登录信息列 -->
        <template #loginInfo="{ record }">
          <div class="login-info-cell">
            <div class="login-time">
              <icon-clock-circle class="info-icon" />
              <span>{{ formatTime(record.last_time) }}</span>
            </div>
            <div v-if="record.last_login_ip" class="login-ip">
              <icon-location class="info-icon" />
              <span>{{ record.last_login_ip }}</span>
            </div>
          </div>
        </template>

        <!-- 注册时间列 -->
        <template #time="{ record }">
          <span class="time-text">{{ formatTime(record.time) }}</span>
        </template>

        <!-- 操作列 -->
        <template #actions="{ record }">
          <div class="action-buttons">
            <a-tooltip content="查看详情">
              <a-button type="text" size="small" @click="handleView(record)">
                <template #icon><icon-eye /></template>
              </a-button>
            </a-tooltip>
            <a-tooltip content="编辑用户">
              <a-button type="text" size="small" @click="handleEdit(record)">
                <template #icon><icon-edit /></template>
              </a-button>
            </a-tooltip>
            <a-dropdown trigger="click">
              <a-button type="text" size="small">
                <template #icon><icon-more /></template>
              </a-button>
              <template #content>
                <a-doption @click="handleAdjustBalance(record)">
                  <template #icon><icon-safe /></template>
                  调整余额
                </a-doption>
                <a-doption @click="handleResetPassword(record)">
                  <template #icon><icon-lock /></template>
                  重置密码
                </a-doption>
                <a-doption @click="handleSetRisk(record)">
                  <template #icon><icon-thunderbolt /></template>
                  设置风控
                </a-doption>
                <a-divider :margin="4" />
                <a-doption class="danger-option" @click="handleDelete(record)">
                  <template #icon><icon-delete /></template>
                  删除用户
                </a-doption>
              </template>
            </a-dropdown>
          </div>
        </template>
      </a-table>
    </a-card>

    <!-- 重置密码弹窗 -->
    <a-modal
      v-model:visible="resetModalVisible"
      title="重置密码"
      :width="420"
      @ok="handleResetSubmit"
    >
      <a-form :model="resetForm" layout="vertical">
        <a-form-item label="用户账号">
          <a-input :model-value="String(resetForm.account)" disabled />
        </a-form-item>
        <a-form-item label="新密码" required>
          <a-input-password
            v-model="resetForm.password"
            placeholder="请输入新密码"
            allow-clear
          >
            <template #append>
              <a-button
                type="text"
                size="small"
                @click="resetForm.password = generatePassword()"
              >
                随机生成
              </a-button>
            </template>
          </a-input-password>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 批量风控弹窗 -->
    <a-modal
      v-model:visible="batchRiskVisible"
      title="批量风控设置"
      :width="420"
      @ok="handleBatchRiskSubmit"
    >
      <a-alert type="info" style="margin-bottom: 16px">
        已选择 <strong>{{ selectedRowKeys.length }}</strong> 个用户
      </a-alert>
      <a-form :model="batchRiskForm" layout="vertical">
        <a-form-item label="风控类型" required>
          <a-radio-group v-model="batchRiskForm.risk" direction="vertical">
            <a-radio :value="0">
              <a-tag color="blue">无</a-tag>
              <span style="margin-left: 8px">正常交易，不进行风控</span>
            </a-radio>
            <a-radio :value="-1">
              <a-tag color="orange">亏损</a-tag>
              <span style="margin-left: 8px">控制用户交易结果为亏损</span>
            </a-radio>
            <a-radio :value="1">
              <a-tag color="green">盈利</a-tag>
              <span style="margin-left: 8px">控制用户交易结果为盈利</span>
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 单个用户风控设置弹窗 -->
    <a-modal
      v-model:visible="singleRiskVisible"
      title="风控设置"
      :width="420"
      @ok="handleSingleRiskSubmit"
    >
      <a-form :model="singleRiskForm" layout="vertical">
        <a-form-item label="用户账号">
          <a-input :model-value="String(singleRiskForm.account)" disabled />
        </a-form-item>
        <a-form-item label="风控类型" required>
          <a-radio-group v-model="singleRiskForm.risk" direction="vertical">
            <a-radio :value="0">
              <a-tag color="blue">无</a-tag>
              <span style="margin-left: 8px">正常交易</span>
            </a-radio>
            <a-radio :value="-1">
              <a-tag color="orange">亏损</a-tag>
              <span style="margin-left: 8px">风控亏损</span>
            </a-radio>
            <a-radio :value="1">
              <a-tag color="green">盈利</a-tag>
              <span style="margin-left: 8px">风控盈利</span>
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 编辑用户弹窗 -->
    <a-modal
      v-model:visible="editModalVisible"
      title="编辑用户"
      :width="680"
      @ok="handleEditSubmit"
    >
      <div class="edit-user-form">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="section-title">
            <icon-user class="section-icon" />
            <span>基本信息</span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="用户ID">
                <a-input :model-value="String(editForm.id)" disabled>
                  <template #prefix><icon-idcard /></template>
                </a-input>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="交易账号" required>
                <a-input v-model="editForm.account_number">
                  <template #prefix><icon-at /></template>
                </a-input>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="用户昵称">
                <a-input v-model="editForm.nickname" placeholder="请输入昵称">
                  <template #prefix><icon-user /></template>
                </a-input>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="用户状态">
                <a-tag
                  :color="editForm.status === 1 ? 'red' : 'green'"
                  size="large"
                >
                  {{ editForm.status === 1 ? '已锁定' : '正常' }}
                </a-tag>
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider :margin="16" />

        <!-- 联系方式 -->
        <div class="form-section">
          <div class="section-title">
            <icon-phone class="section-icon" />
            <span>联系方式</span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="手机号码">
                <a-input v-model="editForm.phone" placeholder="请输入手机号">
                  <template #prefix><icon-phone /></template>
                </a-input>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="邮箱地址">
                <a-input v-model="editForm.email" placeholder="请输入邮箱">
                  <template #prefix><icon-email /></template>
                </a-input>
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider :margin="16" />

        <!-- 风控设置 -->
        <div class="form-section">
          <div class="section-title">
            <icon-safe class="section-icon" />
            <span>风控设置</span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="风控类型" required>
                <a-select v-model="editForm.risk" placeholder="请选择风控类型">
                  <a-option :value="0" label="无">
                    <a-tag color="blue" size="small">无</a-tag>
                    <span style="margin-left: 8px">正常交易</span>
                  </a-option>
                  <a-option :value="-1" label="亏损">
                    <a-tag color="orange" size="small">亏损</a-tag>
                    <span style="margin-left: 8px">风控亏损</span>
                  </a-option>
                  <a-option :value="1" label="盈利">
                    <a-tag color="green" size="small">盈利</a-tag>
                    <span style="margin-left: 8px">风控盈利</span>
                  </a-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item label="用户备注">
                <a-textarea
                  v-model="editForm.user_remark"
                  placeholder="请输入备注信息"
                  :auto-size="{ minRows: 2, maxRows: 4 }"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <a-divider :margin="16" />

        <!-- 安全设置 -->
        <div class="form-section">
          <div class="section-title">
            <icon-lock class="section-icon" />
            <span>安全设置</span>
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="登录密码">
                <a-input-password
                  v-model="editForm.password"
                  placeholder="留空表示不修改"
                  allow-clear
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label=" ">
                <a-button
                  type="outline"
                  size="small"
                  @click="editForm.password = generatePassword()"
                >
                  <template #icon><icon-sync /></template>
                  生成随机密码
                </a-button>
              </a-form-item>
            </a-col>
          </a-row>
        </div>
      </div>
    </a-modal>

    <!-- 余额调整弹窗 -->
    <a-modal
      v-model:visible="adjustBalanceVisible"
      title="调整用户余额"
      :width="560"
      @ok="handleAdjustBalanceSubmit"
      @cancel="resetAdjustBalanceForm"
    >
      <a-form :model="adjustBalanceForm" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="用户账号">
              <a-input
                :model-value="String(adjustBalanceForm.account)"
                disabled
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="用户ID">
              <a-input
                :model-value="String(adjustBalanceForm.user_id)"
                disabled
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider :margin="12" />

        <!-- 钱包类型选择 -->
        <a-form-item label="钱包类型" required>
          <a-radio-group v-model="adjustBalanceForm.wallet_type">
            <a-radio value="spot">
              <a-tag color="arcoblue" size="small">现货账户</a-tag>
            </a-radio>
            <a-radio value="contract">
              <a-tag color="orange" size="small">永续合约账户</a-tag>
            </a-radio>
            <a-radio value="delivery">
              <a-tag color="purple" size="small">交割合约账户</a-tag>
            </a-radio>
          </a-radio-group>
        </a-form-item>

        <!-- 当前余额展示 -->
        <a-alert v-if="userWalletInfo" type="info" style="margin-bottom: 16px">
          <template #icon><icon-info-circle /></template>
          <div class="wallet-info">
            <div class="wallet-title">
              当前 {{ getWalletTypeName(adjustBalanceForm.wallet_type) }} USDT 余额
            </div>
            <div style="margin-top: 12px; text-align: center">
              <div class="balance-item-large">
                <div class="balance-label">可用余额</div>
                <div class="balance-value-large">
                  {{ getCurrentBalance().toFixed(8) }} USDT
                </div>
              </div>
            </div>
          </div>
        </a-alert>

        <!-- 调整参数 -->
        <a-form-item label="调整金额 (USDT)" required>
          <a-input-number
            v-model="adjustBalanceForm.amount"
            placeholder="请输入调整金额"
            :min="0.00000001"
            :precision="8"
            :step="1"
            style="width: 100%"
          >
            <template #prefix>USDT</template>
          </a-input-number>
        </a-form-item>

        <a-form-item label="操作类型" required>
          <a-radio-group v-model="adjustBalanceForm.operation_type">
            <a-radio value="add">
              <a-tag color="green" size="small">充值</a-tag>
            </a-radio>
            <a-radio value="subtract">
              <a-tag color="orange" size="small">扣除</a-tag>
            </a-radio>
          </a-radio-group>
        </a-form-item>

        <a-form-item label="操作原因">
          <div style="margin-bottom: 8px">
            <a-space wrap>
              <a-tag
                color="arcoblue"
                checkable
                style="cursor: pointer"
                @click="adjustBalanceForm.reason = '活动赠送'"
              >
                活动赠送
              </a-tag>
              <a-tag
                color="arcoblue"
                checkable
                style="cursor: pointer"
                @click="adjustBalanceForm.reason = '充值补单'"
              >
                充值补单
              </a-tag>
              <a-tag
                color="arcoblue"
                checkable
                style="cursor: pointer"
                @click="adjustBalanceForm.reason = '客服补偿'"
              >
                客服补偿
              </a-tag>
              <a-tag
                color="arcoblue"
                checkable
                style="cursor: pointer"
                @click="adjustBalanceForm.reason = '违规扣款'"
              >
                违规扣款
              </a-tag>
              <a-tag
                color="arcoblue"
                checkable
                style="cursor: pointer"
                @click="adjustBalanceForm.reason = '后台调整'"
              >
                后台调整
              </a-tag>
            </a-space>
          </div>
          <a-textarea
            v-model="adjustBalanceForm.reason"
            placeholder="请输入操作原因（选填），将记录到账户变动日志中"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            :max-length="200"
            show-word-limit
          />
        </a-form-item>

        <a-alert type="warning">
          <template #icon><icon-exclamation-circle /></template>
          <div>
            <div><strong>操作提醒：</strong></div>
            <ul style="margin: 4px 0 0 0; padding-left: 20px">
              <li>调整金额将直接影响用户可用余额</li>
              <li>系统会自动生成账户变动日志记录</li>
              <li>请谨慎操作，确认无误后再提交</li>
            </ul>
          </div>
        </a-alert>
      </a-form>
    </a-modal>

    <!-- 列设置抽屉 -->
    <a-drawer
      v-model:visible="columnSettingVisible"
      title="列设置"
      :width="320"
    >
      <a-checkbox-group
        v-model="selectedColumns"
        direction="vertical"
        class="column-checkbox-group"
      >
        <a-checkbox
          v-for="col in allColumns"
          :key="col.dataIndex"
          :value="col.dataIndex"
        >
          {{ col.title }}
        </a-checkbox>
      </a-checkbox-group>
      <template #footer>
        <a-button type="primary" @click="columnSettingVisible = false"
          >确定</a-button
        >
        <a-button @click="resetColumns">重置</a-button>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted, h } from 'vue';
  import { useRouter } from 'vue-router';
  import { Message, Modal } from '@arco-design/web-vue';
  import {
    IconSearch,
    IconRefresh,
    IconSync,
    IconDownload,
    IconCheckCircle,
    IconEye,
    IconEdit,
    IconLock,
    IconUnlock,
    IconDown,
    IconUp,
    IconDelete,
    IconUser,
    IconIdcard,
    IconAt,
    IconPhone,
    IconEmail,
    IconSafe,
    IconMore,
    IconMoreVertical,
    IconSettings,
    IconThunderbolt,
    IconClockCircle,
    IconLocation,
    IconInfoCircle,
    IconExclamationCircle,
  } from '@arco-design/web-vue/es/icon';
  import axios from 'axios';
  import dayjs from 'dayjs';

  const router = useRouter();
  const loading = ref(false);
  const showAdvanced = ref(false);
  const selectedRowKeys = ref<number[]>([]);
  const resetModalVisible = ref(false);
  const batchRiskVisible = ref(false);
  const singleRiskVisible = ref(false);
  const editModalVisible = ref(false);
  const columnSettingVisible = ref(false);
  const adjustBalanceVisible = ref(false);

  // 搜索表单
  const searchForm = reactive({
    keyword: '',
    real_name: '',
    status: undefined as number | undefined,
    is_realname: undefined as number | undefined,
    risk: undefined as number | undefined,
  });

  // 表格数据
  const tableData = ref<any[]>([]);

  // 分页配置
  const pagination = reactive({
    current: 1,
    pageSize: 15,
    total: 0,
    showTotal: true,
    showPageSize: true,
    pageSizeOptions: [10, 15, 20, 50],
  });

  const rowSelection = computed(() => ({
    type: 'checkbox' as const,
    showCheckedAll: true,
    selectedRowKeys: selectedRowKeys.value,
    onChange: (keys: (string | number)[]) => {
      selectedRowKeys.value = keys as number[];
    },
  }));

  const resetForm = reactive({ id: '' as any, account: '', password: '' });
  const batchRiskForm = reactive({ risk: 0 });
  const singleRiskForm = reactive({ id: '' as any, account: '', risk: 0 });

  // 余额调整表单
  const adjustBalanceForm = reactive({
    user_id: '' as any,
    account: '',
    wallet_type: 'spot' as 'spot' | 'contract' | 'delivery' | 'fund',
    operation_type: 'add',
    amount: 0,
    reason: '',
  });

  // 用户钱包信息（USDT）
  const userWalletInfo = ref<any>(null);

  const editForm = reactive({
    id: '' as any,
    account_number: '',
    phone: '',
    email: '',
    nickname: '',
    risk: 0,
    user_remark: '',
    password: '',
    status: 0,
  });

  // 所有列定义
  const allColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      slotName: 'id',
      width: 70,
      align: 'center' as const,
    },
    {
      title: '用户信息',
      dataIndex: 'userInfo',
      slotName: 'userInfo',
      width: 200,
    },
    {
      title: '联系方式',
      dataIndex: 'contact',
      slotName: 'contact',
      width: 180,
    },
    {
      title: '现货USDT',
      dataIndex: 'spotBalance',
      slotName: 'spotBalance',
      width: 130,
      align: 'right' as const,
    },
    {
      title: '永续合约USDT',
      dataIndex: 'contractBalance',
      slotName: 'contractBalance',
      width: 140,
      align: 'right' as const,
    },
    {
      title: '交割合约USDT',
      dataIndex: 'deliveryBalance',
      slotName: 'deliveryBalance',
      width: 140,
      align: 'right' as const,
    },
    {
      title: '资金钱包USDT',
      dataIndex: 'fundBalance',
      slotName: 'fundBalance',
      width: 140,
      align: 'right' as const,
    },
    {
      title: '状态',
      dataIndex: 'status',
      slotName: 'status',
      width: 90,
      align: 'center' as const,
    },
    {
      title: '风控',
      dataIndex: 'risk',
      slotName: 'risk',
      width: 80,
      align: 'center' as const,
    },
    { title: '邀请关系', dataIndex: 'invite', slotName: 'invite', width: 150 },
    {
      title: '最近登录',
      dataIndex: 'loginInfo',
      slotName: 'loginInfo',
      width: 160,
    },
    { title: '注册时间', dataIndex: 'time', slotName: 'time', width: 150 },
    {
      title: '操作',
      dataIndex: 'actions',
      slotName: 'actions',
      width: 140,
      fixed: 'right' as const,
      align: 'center' as const,
    },
  ];

  // 默认显示的列
  const defaultColumns = [
    'id',
    'userInfo',
    'contact',
    'spotBalance',
    'contractBalance',
    'deliveryBalance',
    'fundBalance',
    'status',
    'risk',
    'loginInfo',
    'time',
    'actions',
  ];
  const selectedColumns = ref<string[]>([...defaultColumns]);

  // 可见列
  const visibleColumns = computed(() => {
    return allColumns.filter((col) =>
      selectedColumns.value.includes(col.dataIndex)
    );
  });

  // 表格滚动宽度
  const tableScrollWidth = computed(() => {
    return visibleColumns.value.reduce(
      (sum, col) => sum + (col.width || 100),
      0
    );
  });

  // 格式化时间
  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm');
  };

  const isRealVerified = (record: any) =>
    record?.is_realname === 2 || record?.is_real === 1;

  const getRiskText = (risk?: number) => {
    if (risk === -1) return '亏损';
    if (risk === 1) return '盈利';
    return '无';
  };

  const getRiskColor = (risk?: number) => {
    if (risk === -1) return 'orange';
    if (risk === 1) return 'green';
    return 'blue';
  };

  const generatePassword = () => {
    const charset = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789';
    return Array.from(
      { length: 10 },
      () => charset[Math.floor(Math.random() * charset.length)]
    ).join('');
  };

  // 获取用户列表
  const fetchUserList = async () => {
    try {
      loading.value = true;
      const params = {
        page: pagination.current,
        page_size: pagination.pageSize,
        keyword: searchForm.keyword || undefined,
        status: searchForm.status,
        is_real: searchForm.is_realname,
        real_name: searchForm.real_name || undefined,
        risk: searchForm.risk,
      };
      const response = await axios.post('/admin/user/list', params);
      if (response.data?.data) {
        tableData.value = response.data.data.list || [];
        pagination.total = response.data.data.total || 0;
      }
    } catch (error: any) {
      Message.error(error.message || '获取用户列表失败');
    } finally {
      loading.value = false;
    }
  };

  const handleSearch = () => {
    pagination.current = 1;
    selectedRowKeys.value = [];
    fetchUserList();
  };

  const handleReset = () => {
    Object.assign(searchForm, {
      keyword: '',
      real_name: '',
      status: undefined,
      is_realname: undefined,
      risk: undefined,
    });
    pagination.current = 1;
    selectedRowKeys.value = [];
    fetchUserList();
  };

  const handleRefresh = () => fetchUserList();
  const handlePageChange = (page: number) => {
    pagination.current = page;
    fetchUserList();
  };
  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.current = 1;
    fetchUserList();
  };

  const handleView = (record: any) =>
    router.push(`/admin/user/detail/${record.id}`);
  const handleViewParent = (record: any) =>
    router.push(`/admin/user/detail/${record.parent_id}`);

  // 快速切换状态
  const handleQuickToggleStatus = async (record: any, isActive: boolean) => {
    try {
      await axios.post(`/admin/user/${isActive ? 'activate' : 'freeze'}`, {
        id: record.id,
      });
      Message.success(isActive ? '已解锁' : '已锁定');
      fetchUserList();
    } catch (error: any) {
      Message.error(error.message || '操作失败');
    }
  };

  // 编辑用户
  const handleEdit = (record: any) => {
    Object.assign(editForm, {
      id: record.id,
      account_number: record.account_number || '',
      phone: record.phone || '',
      email: record.email || '',
      nickname: record.nickname || '',
      risk: record.risk ?? 0,
      user_remark: record.user_remark || '',
      password: '',
      status: record.status ?? 0,
    });
    editModalVisible.value = true;
  };

  const handleEditSubmit = async () => {
    try {
      const payload: Record<string, unknown> = {
        account_number: editForm.account_number,
        phone: editForm.phone,
        email: editForm.email,
        nickname: editForm.nickname,
        risk: editForm.risk,
        user_remark: editForm.user_remark,
      };
      if (editForm.password) payload.password = editForm.password;
      await axios.put(`/admin/user/${editForm.id}`, payload);
      Message.success('用户信息已更新');
      editModalVisible.value = false;
      fetchUserList();
    } catch (error: any) {
      Message.error(error.message || '更新失败');
    }
  };

  // 重置密码
  const handleResetPassword = (record: any) => {
    resetForm.id = record.id;
    resetForm.account = record.account_number;
    resetForm.password = generatePassword();
    resetModalVisible.value = true;
  };

  const handleResetSubmit = async () => {
    if (!resetForm.password) {
      Message.warning('请输入新密码');
      return;
    }
    try {
      await axios.post('/admin/user/reset-password', {
        id: resetForm.id,
        password: resetForm.password,
      });
      Message.success('密码重置成功');
      resetModalVisible.value = false;
    } catch (error: any) {
      Message.error(error.message || '重置密码失败');
    }
  };

  // 快速设置风控
  const handleQuickSetRisk = async (record: any, risk: any) => {
    try {
      await axios.post('/admin/user/batch-risk', {
        user_ids: [record.id],
        risk_level: Number(risk),
      });
      Message.success('风控设置成功');
      record.risk = Number(risk);
    } catch (error: any) {
      Message.error(error.message || '设置失败');
    }
  };

  // 设置单个用户风控
  const handleSetRisk = (record: any) => {
    singleRiskForm.id = record.id;
    singleRiskForm.account = record.account_number;
    singleRiskForm.risk = record.risk ?? 0;
    singleRiskVisible.value = true;
  };

  const handleSingleRiskSubmit = async () => {
    try {
      await axios.post('/admin/user/batch-risk', {
        user_ids: [singleRiskForm.id],
        risk_level: singleRiskForm.risk,
      });
      Message.success('风控设置成功');
      singleRiskVisible.value = false;
      fetchUserList();
    } catch (error: any) {
      Message.error(error.message || '设置失败');
    }
  };

  // 批量风控
  const handleBatchRisk = () => {
    if (selectedRowKeys.value.length === 0) {
      Message.warning('请先选择用户');
      return;
    }
    batchRiskForm.risk = 0;
    batchRiskVisible.value = true;
  };

  const handleBatchRiskSubmit = async () => {
    try {
      await axios.post('/admin/user/batch-risk', {
        user_ids: selectedRowKeys.value,
        risk_level: batchRiskForm.risk,
      });
      Message.success('批量风控设置成功');
      batchRiskVisible.value = false;
      selectedRowKeys.value = [];
      fetchUserList();
    } catch (error: any) {
      Message.error(error.message || '批量设置失败');
    }
  };

  // 批量锁定/解锁
  const handleBatchStatus = (status: number) => {
    if (selectedRowKeys.value.length === 0) {
      Message.warning('请先选择用户');
      return;
    }
    const action = status === 1 ? '锁定' : '解锁';
    Modal.confirm({
      title: `批量${action}`,
      content: `确定要${action}选中的 ${selectedRowKeys.value.length} 个用户吗？`,
      onOk: async () => {
        try {
          const endpoint =
            status === 1 ? '/admin/user/freeze' : '/admin/user/activate';
          await Promise.all(
            selectedRowKeys.value.map((id) => axios.post(endpoint, { id }))
          );
          Message.success(`批量${action}成功`);
          selectedRowKeys.value = [];
          fetchUserList();
        } catch (error: any) {
          Message.error(error.message || `批量${action}失败`);
        }
      },
    });
  };

  // 删除用户
  const handleDelete = (record: any) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除用户 ${record.account_number} 吗？此操作不可恢复！`,
      okButtonProps: { status: 'danger' },
      onOk: async () => {
        try {
          await axios.delete(`/admin/user/${record.id}`);
          Message.success('删除成功');
          fetchUserList();
        } catch (error: any) {
          Message.error(error.message || '删除失败');
        }
      },
    });
  };

  // 余额调整
  const handleAdjustBalance = async (record: any) => {
    // 重置表单
    Object.assign(adjustBalanceForm, {
      user_id: record.id,
      account: record.account_number,
      operation_type: 'add',
      amount: 0,
      reason: '',
    });

    // 获取用户 USDT 钱包信息
    try {
      const response = await axios.get(`/admin/user/${record.id}/wallets`);
      if (response.data?.data) {
        // 查找 USDT 钱包（currency_id = 3 是 USDT）
        const usdtWallet = response.data.data.find(
          (w: any) => w.currency === 3 || w.currency_id === 3
        );
        userWalletInfo.value = usdtWallet || null;
      }
    } catch (error: any) {
      Message.warning('获取用户钱包信息失败');
      userWalletInfo.value = null;
    }

    adjustBalanceVisible.value = true;
  };

  // 快速调整余额（点击余额时触发）
  const handleQuickAdjustBalance = (
    record: any,
    walletType: 'spot' | 'contract' | 'delivery' | 'fund' = 'spot'
  ) => {
    // 直接使用列表中的余额数据，无需重新查询
    Object.assign(adjustBalanceForm, {
      user_id: record.id,
      account: record.account_number,
      wallet_type: walletType,
      operation_type: 'add',
      amount: 0,
      reason: '',
    });

    // 设置钱包信息（直接使用列表数据）
    userWalletInfo.value = {
      currency_id: 3,
      spot_balance: record.spot_balance || 0,
      contract_balance: record.contract_balance || 0,
      delivery_balance: record.delivery_balance || 0,
      fund_balance: record.fund_balance || 0,
    };

    adjustBalanceVisible.value = true;
  };

  // 获取钱包类型名称
  const getWalletTypeName = (type: string) => {
    if (type === 'spot') return '现货';
    if (type === 'contract') return '永续合约';
    if (type === 'delivery') return '交割合约';
    if (type === 'fund') return '资金钱包';
    return '未知';
  };

  // 获取钱包类型完整名称（带“账户”）
  const getWalletTypeFullName = (type: string) => {
    if (type === 'spot') return '现货账户';
    if (type === 'contract') return '永续合约账户';
    if (type === 'delivery') return '交割合约账户';
    if (type === 'fund') return '资金钱包';
    return '未知账户';
  };

  // 获取当前余额
  const getCurrentBalance = () => {
    if (!userWalletInfo.value) return 0;
    const type = adjustBalanceForm.wallet_type;
    if (type === 'spot') return userWalletInfo.value.spot_balance || 0;
    if (type === 'contract') return userWalletInfo.value.contract_balance || 0;
    if (type === 'delivery') return userWalletInfo.value.delivery_balance || 0;
    if (type === 'fund') return userWalletInfo.value.fund_balance || 0;
    return 0;
  };

  const resetAdjustBalanceForm = () => {
    Object.assign(adjustBalanceForm, {
      user_id: 0,
      account: '',
      wallet_type: 'spot',
      operation_type: 'add',
      amount: 0,
      reason: '',
    });
    userWalletInfo.value = null;
  };

  const handleAdjustBalanceSubmit = async () => {
    // 表单验证
    if (!adjustBalanceForm.amount || adjustBalanceForm.amount <= 0) {
      Message.warning('请输入正确的调整金额');
      return;
    }
    // 操作原因改为可选，不再强制验证

    // 计算实际金额（扣除时为负数）
    const actualAmount =
      adjustBalanceForm.operation_type === 'add'
        ? adjustBalanceForm.amount
        : -adjustBalanceForm.amount;

    // 确认弹窗
    const operationText =
      adjustBalanceForm.operation_type === 'add' ? '充值' : '扣除';
    const walletTypeText = getWalletTypeFullName(adjustBalanceForm.wallet_type);

    Modal.confirm({
      title: `确认${operationText}余额`,
      content: () =>
        h('div', { style: { lineHeight: '1.8' } }, [
          h('p', { style: { marginBottom: '8px' } }, [
            h('strong', '用户账号：'),
            adjustBalanceForm.account,
          ]),
          h('p', { style: { marginBottom: '8px' } }, [
            h('strong', '账户类型：'),
            walletTypeText,
          ]),
          h('p', { style: { marginBottom: '8px' } }, [
            h('strong', '操作类型：'),
            h(
              'span',
              {
                style: {
                  color:
                    adjustBalanceForm.operation_type === 'add'
                      ? '#00b42a'
                      : '#ff7d00',
                  fontWeight: '500',
                },
              },
              operationText
            ),
          ]),
          h('p', { style: { marginBottom: '8px' } }, [
            h('strong', '调整金额：'),
            `${adjustBalanceForm.amount} USDT`,
          ]),
          h('p', { style: { marginBottom: '12px' } }, [
            h('strong', '操作原因：'),
            adjustBalanceForm.reason || '无',
          ]),
          h(
            'p',
            {
              style: {
                color: '#f53f3f',
                marginTop: '12px',
                padding: '8px 12px',
                background: '#fff1f0',
                borderRadius: '4px',
                fontSize: '13px',
              },
            },
            '此操作将直接影响用户余额，请确认后操作！'
          ),
        ]),
      okText: '确认提交',
      cancelText: '取消',
      onOk: async () => {
        try {
          // 根据钱包类型调用不同的API
          if (adjustBalanceForm.wallet_type === 'fund') {
            // 资金钱包使用专用API
            await axios.post('/admin/wallet/fund/adjust', {
              user_id: adjustBalanceForm.user_id,
              amount: adjustBalanceForm.amount.toString(),
              type: adjustBalanceForm.operation_type, // 'add' 或 'subtract'
              remark: adjustBalanceForm.reason
                ? adjustBalanceForm.reason.trim()
                : '',
            });
          } else {
            // 现货、合约、交割钱包使用原API
            await axios.post('/admin/user/adjust-balance', {
              user_id: adjustBalanceForm.user_id,
              currency_id: 1,
              wallet_type: adjustBalanceForm.wallet_type,
              amount: actualAmount,
              reason: adjustBalanceForm.reason
                ? adjustBalanceForm.reason.trim()
                : '',
            });
          }

          Message.success(`${operationText}成功！已记录账户变动日志`);
          adjustBalanceVisible.value = false;
          resetAdjustBalanceForm();
          fetchUserList();
        } catch (error: any) {
          Message.error(
            error.response?.data?.error ||
              error.response?.data?.message ||
              error.message ||
              `${operationText}失败`
          );
        }
      },
    });
  };

  // 导出
  const handleExport = () => {
    const params = new URLSearchParams();
    if (searchForm.keyword) params.append('account', searchForm.keyword);
    if (searchForm.status !== undefined)
      params.append('status', String(searchForm.status));
    if (searchForm.is_realname !== undefined)
      params.append('is_real', String(searchForm.is_realname));
    const baseURL = import.meta.env.VITE_API_BASE_URL || '';
    window.open(`${baseURL}/admin/user/export?${params.toString()}`);
  };

  // 列设置
  const handleColumnSetting = () => {
    columnSettingVisible.value = true;
  };
  const resetColumns = () => {
    selectedColumns.value = [...defaultColumns];
  };

  onMounted(() => fetchUserList());
</script>

<style scoped lang="less">
  .user-list-container {
    padding: 16px;
    background: var(--color-bg-1);
    min-height: 100%;
  }

  .search-card {
    margin-bottom: 16px;

    :deep(.arco-card-body) {
      padding: 16px 20px;
    }

    .search-form {
      :deep(.arco-form-item) {
        margin-bottom: 0;
        margin-right: 12px;
      }
    }

    .search-actions {
      display: flex;
      justify-content: flex-end;
      align-items: center;
    }
  }

  .table-card {
    :deep(.arco-card-body) {
      padding: 16px 20px;
    }
  }

  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      .table-title {
        font-size: 16px;
        font-weight: 600;
        color: var(--color-text-1);
      }
    }
  }

  .user-table {
    :deep(.arco-table-th) {
      background: var(--color-fill-1);
      font-weight: 500;
    }

    :deep(.arco-table-tr:hover) {
      .arco-table-td {
        background: var(--color-fill-1);
      }
    }
  }

  .cell-id {
    font-family: 'Monaco', monospace;
    font-size: 13px;
    color: var(--color-text-2);
  }

  .user-info-cell {
    .user-main {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-bottom: 4px;

      .account {
        font-weight: 500;
        color: var(--color-text-1);
      }

      .verify-tag {
        font-size: 11px;
        line-height: 1;
      }
    }

    .user-sub {
      display: flex;
      gap: 8px;
      font-size: 12px;
      color: var(--color-text-3);

      .realname::before {
        content: '·';
        margin-right: 8px;
      }
    }
  }

  .contact-cell {
    .contact-item {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 13px;
      margin-bottom: 2px;

      .contact-icon {
        color: var(--color-text-3);
        font-size: 12px;
      }

      .email-text {
        max-width: 140px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .empty-text {
      color: var(--color-text-4);
    }
  }

  // USDT余额列样式
  .balance-cell {
    padding: 4px 0;
    transition: all 0.2s;

    &:hover {
      .balance-amount {
        background: var(--color-fill-2);
        transform: scale(1.02);
      }
    }

    .balance-amount {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      padding: 6px 12px;
      background: var(--color-fill-1);
      border-radius: 6px;
      transition: all 0.2s;
      border: 1px solid transparent;

      &:hover {
        border-color: rgb(var(--primary-6));
        box-shadow: 0 2px 8px rgba(var(--primary-6), 0.15);
      }

      .balance-value {
        font-size: 15px;
        font-weight: 600;
        color: var(--color-text-1);
        font-family: 'Monaco', 'Consolas', monospace;
        letter-spacing: 0.5px;
      }

      .balance-unit {
        font-size: 12px;
        color: var(--color-text-3);
        font-weight: 500;
      }
    }
  }

  .risk-tag {
    min-width: 50px;
    justify-content: center;
  }

  .invite-cell {
    font-size: 12px;

    .label {
      color: var(--color-text-3);
      margin-right: 4px;
    }

    .parent-info {
      margin-top: 4px;
    }
  }

  .login-info-cell {
    font-size: 12px;

    .info-icon {
      color: var(--color-text-3);
      margin-right: 4px;
      font-size: 12px;
    }

    .login-ip {
      margin-top: 2px;
      color: var(--color-text-3);
    }
  }

  .time-text {
    font-size: 13px;
    color: var(--color-text-2);
  }

  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 2px;
  }

  .danger-option {
    color: rgb(var(--red-6)) !important;
  }

  .edit-user-form {
    .form-section {
      .section-title {
        display: flex;
        align-items: center;
        margin-bottom: 16px;
        font-size: 14px;
        font-weight: 500;
        color: var(--color-text-1);

        .section-icon {
          margin-right: 8px;
          color: rgb(var(--primary-6));
          font-size: 16px;
        }
      }
    }

    :deep(.arco-form-item) {
      margin-bottom: 16px;
    }
  }

  .column-checkbox-group {
    :deep(.arco-checkbox) {
      margin-bottom: 12px;
    }
  }

  // 余额调整弹窗样式
  .wallet-info {
    .wallet-title {
      font-size: 13px;
      font-weight: 500;
      color: var(--color-text-1);
      margin-bottom: 8px;
    }

    .balance-item-large {
      padding: 20px;
      background: linear-gradient(
        135deg,
        rgb(var(--primary-6)) 0%,
        rgb(var(--primary-5)) 100%
      );
      border-radius: 8px;

      .balance-label {
        font-size: 14px;
        color: rgba(255, 255, 255, 0.9);
        margin-bottom: 8px;
      }

      .balance-value-large {
        font-size: 28px;
        font-weight: 700;
        color: white;
        font-family: 'Monaco', monospace;
        letter-spacing: 1px;
      }
    }

    .balance-item {
      text-align: center;
      padding: 8px;
      background: var(--color-fill-1);
      border-radius: 4px;

      .balance-label {
        font-size: 12px;
        color: var(--color-text-3);
        margin-bottom: 4px;
      }

      .balance-value {
        font-size: 14px;
        font-weight: 600;
        color: rgb(var(--primary-6));
        font-family: 'Monaco', monospace;
      }
    }

    .balance-item-small {
      text-align: center;
      padding: 4px;
      background: var(--color-fill-2);
      border-radius: 4px;

      .balance-label {
        font-size: 11px;
        color: var(--color-text-4);
        margin-bottom: 2px;
      }

      .balance-value {
        font-size: 12px;
        font-weight: 500;
        color: var(--color-text-2);
        font-family: 'Monaco', monospace;
      }
    }
  }
</style>
