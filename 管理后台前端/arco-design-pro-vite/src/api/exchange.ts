import axios from 'axios';

// ===== 统计数据 =====
export interface DashboardData {
  total_users: number;
  today_new_users: number;
  total_deposit: number;
  total_withdraw: number;
  pending_withdraw: number;
  pending_kyc: number;
  today_trades: number;
  today_trade_amount: number;
}

export interface UserStatistics {
  total: number;
  active: number;
  frozen: number;
  verified: number;
  unverified: number;
}

export interface TradeStatistics {
  spot_total: number;
  spot_amount: number;
  lever_total: number;
  lever_amount: number;
  micro_total: number;
  micro_amount: number;
}

export function getDashboardData() {
  return axios.get<DashboardData>('/admin/statistics/dashboard');
}

export function getUserStatistics() {
  return axios.get<UserStatistics>('/admin/statistics/user');
}

export function getTradeStatistics() {
  return axios.get<TradeStatistics>('/admin/statistics/trade');
}

// ===== 币种管理 =====
export interface Currency {
  id: number;
  name: string;
  logo: string;
  type: string;
  decimal_scale: number;
  rate: number;
  is_legal: number;
  is_lever: number;
  is_micro: number;
  is_display: number;
  is_change: number;
  sort: number;
  create_time: number;
  update_time: number;
  // 风控相关字段
  risk_prob_enabled: number;
  risk_profit_probability: number;
  risk_money_enabled: number;
  risk_money_min: number;
  risk_money_max: number;
  risk_money_result: number;
  risk_time_enabled: number;
  risk_time_start: string;
  risk_time_end: string;
  risk_time_result: number;
}

export function getCurrencyList(params?: any) {
  return axios.get<Currency[]>('/currency/list', { params });
}

export function createCurrency(data: Partial<Currency>) {
  return axios.post('/currency/create', data);
}

export function updateCurrency(id: number, data: Partial<Currency>) {
  return axios.put(`/currency/${id}`, data);
}

export function deleteCurrency(id: number) {
  return axios.delete(`/currency/${id}`);
}

// ===== 交易对管理 =====
export interface CurrencyMatch {
  id: number;
  currency_id: number;
  legal_id: number;
  symbol: string;
  is_display: number;
  is_trade: number;
  change_fee: number;
  legal_fee: number;
  lever_share_num: number;
  spread: number;
  overnight: number;
  lever_trade_fee: number;
  lever_min_share: number;
  lever_max_share: number;
  create_time: number;
}

export function getMatchList(params?: any) {
  return axios.get<CurrencyMatch[]>('/match/list', { params });
}

export function createMatch(data: Partial<CurrencyMatch>) {
  return axios.post('/match/create', data);
}

export function updateMatch(id: number, data: Partial<CurrencyMatch>) {
  return axios.put(`/match/${id}`, data);
}

// ===== 杠杆倍数管理 =====
export interface LeverMultiple {
  id: number;
  type: string;
  value: string;
}

export function getLeverMultiples() {
  return axios.get<LeverMultiple[]>('/admin/lever/multiples');
}

export function createLeverMultiple(data: { type: string; value: string }) {
  return axios.post('/admin/lever/multiple', data);
}

export function updateLeverMultiple(
  id: number,
  data: { type: string; value: string }
) {
  return axios.put(`/admin/lever/multiple/${id}`, data);
}

export function deleteLeverMultiple(id: number) {
  return axios.delete(`/admin/lever/multiple/${id}`);
}

// ===== 用户管理 =====
export interface User {
  id: number;
  phone: string;
  email: string;
  account_number: string;
  status: number;
  account_type: number;
  invitation_code: string;
  parent_id: number;
  create_time: number;
}

export function getUserList(params: {
  page: number;
  page_size: number;
  keyword?: string;
  status?: number;
}) {
  return axios.get<{
    list: User[];
    total: number;
  }>('/admin/users', { params });
}

export function updateUserStatus(id: number, status: number) {
  return axios.put(`/admin/user/${id}/status`, { status });
}

// ===== 钱包管理 =====
export interface Wallet {
  id: number;
  user_id: number;
  currency: number;
  address: string;
  legal_balance: number;
  change_balance: number;
  lever_balance: number;
  micro_balance: number;
  lock_balance: number;
  lock_change: number;
  lock_lever: number;
  status: number;
  create_time: number;
}

export function getWalletList(params: {
  user_id?: number;
  currency_id?: number;
  page: number;
  page_size: number;
}) {
  return axios.get<{
    list: Wallet[];
    total: number;
  }>('/admin/wallets', { params });
}

// ===== 杠杆交易订单管理 =====
export interface LeverOrder {
  id: number;
  user_id: number;
  currency_id: number;
  legal_id: number;
  type: number;
  multiple: number;
  share: number;
  price: number;
  number: number;
  stop_loss_price: number;
  target_profit_price: number;
  current_price: number;
  caution_money: number;
  trade_fee: number;
  overnight_money: number;
  status: number;
  fact_profits: number;
  create_time: number;
  complete_time: number;
}

export function getLeverOrders(params: {
  user_id?: number;
  currency_id?: number;
  status?: number;
  page: number;
  page_size: number;
}) {
  return axios.get<{
    list: LeverOrder[];
    total: number;
  }>('/admin/lever/orders', { params });
}

export function closeLeverOrder(id: number, closePrice?: number) {
  return axios.post(`/admin/lever/close/${id}`, { close_price: closePrice });
}

// ===== 系统设置 =====
export interface Setting {
  id: number;
  key: string;
  value: string;
  name: string;
}

export function getSettings() {
  return axios.get<Setting[]>('/admin/settings');
}

export function updateSetting(key: string, value: string) {
  return axios.put('/admin/setting', { key, value });
}

// ===== 统计数据 =====
export interface DashboardStats {
  total_users: number;
  active_users_today: number;
  total_volume_24h: number;
  total_orders_today: number;
  total_deposits_today: number;
  total_withdrawals_today: number;
}

export function getDashboardStats() {
  return axios.get<DashboardStats>('/admin/stats/dashboard');
}
