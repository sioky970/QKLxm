import axios from 'axios';

// ===== 通用类型 =====

export interface PageParams {
  page: number;
  page_size: number;
}

export interface PageResponse<T> {
  list: T[];
  total: number;
}

// ===== 现货交易 =====

export interface SpotOrder {
  id: number;
  from_user_id: number;
  to_user_id?: number;
  currency: number;
  legal: number;
  type: number;
  price: number;
  number: number;
  deal: number;
  fee: number;
  status: number;
  time: number;
}

export interface SpotListParams extends PageParams {
  account_number?: string;
  status?: number;
}

export function getSpotOrderList(params: SpotListParams) {
  return axios.post<PageResponse<SpotOrder>>(
    '/admin/transaction/spot/list',
    params
  );
}

export function cancelSpotOrder(id: number) {
  return axios.post('/admin/transaction/cancel', { id });
}

// ===== 永续合约 =====

export interface LeverOrder {
  id: number;
  user_id: number;
  account_number?: string;
  currency: number;
  currency_name?: string;
  legal: number;
  legal_name?: string;
  type: number;
  multiple: number;
  share: number;
  price: number;
  current_price: number;
  update_price: number;
  caution_money: number;
  fact_profits: number;
  status: number;
  create_time: number;
  complete_time?: number;
}

export interface LeverListParams extends PageParams {
  account_number?: string;
  currency_id?: number;
  legal_id?: number;
  status?: number;
  type?: string;
}

export function getLeverOrderList(params: LeverListParams) {
  return axios.post<PageResponse<LeverOrder>>('/admin/lever/list', params);
}

export function getLeverOrderDetail(id: number) {
  return axios.get<{ order: LeverOrder; user: unknown }>(`/admin/lever/${id}`);
}

export function closeLeverOrder(id: number, closePrice?: number) {
  return axios.post('/admin/lever/close', { id, close_price: closePrice });
}

export function exportLeverOrders() {
  return axios.get('/admin/lever/export', { responseType: 'blob' });
}

// ===== 交割合约 =====

export interface MicroOrder {
  id: number;
  user_id: number;
  account_number?: string;
  currency_id: number;
  currency_name?: string;
  type: number;
  seconds: number;
  number: number;
  profit_ratio: number;
  open_price: number;
  end_price?: number;
  fact_profits: number;
  profit_result: number;
  pre_profit_result?: number;
  status: number;
  created_at: string;
}

export interface MicroListParams extends PageParams {
  account_number?: string;
  currency_id?: number;
  status?: number;
  result?: string;
}

export function getMicroOrderList(params: MicroListParams) {
  return axios.post<PageResponse<MicroOrder>>(
    '/admin/micro/order/list',
    params
  );
}

export function getMicroOrderDetail(id: number) {
  return axios.get<{ order: MicroOrder; user: unknown }>(
    `/admin/micro/order/${id}`
  );
}

export function updateMicroOrder(
  id: number,
  data: { pre_profit_result?: number; end_price?: number }
) {
  return axios.put(`/admin/micro/order/${id}`, data);
}

export function batchMicroRisk(orderIds: number[], preProfitResult: number) {
  return axios.post('/admin/micro/order/batch-risk', {
    order_ids: orderIds,
    pre_profit_result: preProfitResult,
  });
}
