import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_SYSTEM: AppRouteRecordRaw = {
  path: '/admin/system',
  name: 'SystemManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '系统管理',
    requiresAuth: true,
    icon: 'icon-settings',
    order: 10,
  },
  children: [
    {
      path: 'kyc',
      name: 'KYCManagement',
      component: () => import('@/views/admin/kyc/list.vue'),
      meta: {
        locale: '实名认证',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'finance',
      name: 'FinanceManagement',
      component: () => import('@/views/admin/finance/account-logs.vue'),
      meta: {
        locale: '财务流水',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'system-config',
      name: 'SystemConfig',
      component: () => import('@/views/admin/settings/system.vue'),
      meta: {
        locale: '系统设置',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'admin',
      name: 'AdminManagement',
      component: () => import('@/views/admin/admin/list.vue'),
      meta: {
        locale: '管理员管理',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'news',
      name: 'NewsManagement',
      component: () => import('@/views/admin/news/list.vue'),
      meta: {
        locale: '新闻管理',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'deposit-address',
      name: 'DepositAddressManagement',
      component: () => import('@/views/admin/settings/deposit-address.vue'),
      meta: {
        locale: '充值地址管理',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default ADMIN_SYSTEM;
