import axios from 'axios';

// ===== 充值地址管理 =====

export interface DepositAddress {
  id: number;
  network: string;
  address: string;
  qr_code: string;
  status: number;
  sort: number;
  create_time: number;
  update_time: number;
}

export interface CreateDepositAddressRequest {
  network: string;
  address: string;
  qr_code?: string;
  status?: number;
  sort?: number;
}

export interface UpdateDepositAddressRequest {
  address?: string;
  qr_code?: string;
  status?: number;
  sort?: number;
}

// 获取充值地址列表
export function getDepositAddressList() {
  return axios.get<DepositAddress[]>('/admin/deposit/address/list');
}

// 创建充值地址
export function createDepositAddress(data: CreateDepositAddressRequest) {
  return axios.post<DepositAddress>('/admin/deposit/address/create', data);
}

// 更新充值地址
export function updateDepositAddress(
  id: number,
  data: UpdateDepositAddressRequest
) {
  return axios.put(`/admin/deposit/address/${id}`, data);
}

// 删除充值地址
export function deleteDepositAddress(id: number) {
  return axios.delete(`/admin/deposit/address/${id}`);
}

// 更新充值地址状态
export function updateDepositAddressStatus(id: number, status: number) {
  return axios.put(`/admin/deposit/address/${id}`, { status });
}

// ===== 充值订单审核 =====

export interface DepositOrder {
  id: number;
  order_no: string;
  user_id: number;
  account_number?: string;
  network: string;
  address: string;
  amount: number;
  screenshot: string;
  status: number; // 0:待审核 1:已通过 2:已拒绝 3:已取消 4:已过期
  admin_id?: number;
  admin_remark?: string;
  create_time: number;
  update_time: number;
  review_time?: number;
}

export interface DepositOrderListRequest {
  user_id?: number;
  order_no?: string;
  network?: string;
  status?: number;
  page: number;
  page_size: number;
}

export interface DepositOrderListResponse {
  data: DepositOrder[];
  count: number;
}

// 获取充值订单列表
export function getDepositOrderList(params: DepositOrderListRequest) {
  return axios.post<DepositOrderListResponse>('/admin/deposit/order/list', params);
}

// 获取充值订单详情
export function getDepositOrderDetail(id: number) {
  return axios.get<DepositOrder>(`/admin/deposit/order/${id}`);
}

export interface ApproveDepositRequest {
  order_id: number;
  remark?: string;
}

// 审核通过充值订单
export function approveDepositOrder(data: ApproveDepositRequest) {
  return axios.post('/admin/deposit/approve', data);
}

export interface RejectDepositRequest {
  order_id: number;
  remark?: string;
}

// 审核拒绝充值订单
export function rejectDepositOrder(data: RejectDepositRequest) {
  return axios.post('/admin/deposit/reject', data);
}
