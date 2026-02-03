import axios from 'axios';
import type { RouteRecordNormalized } from 'vue-router';
import { UserState } from '@/store/modules/user/types';

export interface LoginData {
  username: string;
  password: string;
}

export interface LoginRes {
  token: string;
  admin: {
    id: number;
    username: string;
    role_id: number;
    is_super: number;
  };
}

// 管理员登录
export function login(data: LoginData) {
  return axios.post<LoginRes>('/admin/login', data);
}

// 管理员退出
export function logout() {
  return axios.post('/admin/logout');
}

// 获取管理员信息
export function getUserInfo() {
  return axios.get<UserState>('/admin/info');
}

export function getMenuList() {
  return axios.post<RouteRecordNormalized[]>('/user/menu');
}
