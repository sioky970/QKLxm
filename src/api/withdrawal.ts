import axios from 'axios';

// ===== 提现管理 =====

export interface WithdrawalItem {
  id: number;
  user_id: number;
  currency_id: number;
  address: string;
  number: number;
  create_time: number;
  rate: number;
  status: number;
  notes: string;
  real_number: number;
  txid: string;
  update_time: number;
  to_adddress: string;
  usdt_type: string;
  account_number?: string;
  currency_name?: string;
  // 区块链提现相关字段
  withdraw_type?: number; // 1=银行转账, 2=区块链提现
  network_type?: string; // 网络类型: TRC20/ERC20/BEP20等
  chain_address?: string; // 区块链钱包地址
}

export interface WithdrawalListRequest {
  page: number;
  page_size: number;
  account_number?: string;
  status?: number;
  currency_id?: number;
}

export interface WithdrawalListResponse {
  list: WithdrawalItem[];
  total: number;
}

export function getWithdrawalList(params: WithdrawalListRequest) {
  return axios.post<WithdrawalListResponse>('/admin/withdrawal/list', params);
}

export function getWithdrawalDetail(id: number) {
  return axios.get<WithdrawalItem>(`/admin/withdrawal/${id}`);
}

export interface ApproveWithdrawalRequest {
  id: number;
  method?: string;
  txid?: string;
  notes?: string;
}

export function approveWithdrawal(data: ApproveWithdrawalRequest) {
  return axios.post('/admin/withdrawal/approve', data);
}

export interface RejectWithdrawalRequest {
  id: number;
  reason?: string;
}

export function rejectWithdrawal(data: RejectWithdrawalRequest) {
  return axios.post('/admin/withdrawal/reject', data);
}
