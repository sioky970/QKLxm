import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_TRANSACTION: AppRouteRecordRaw = {
  path: '/admin/transaction',
  name: 'TransactionManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '交易管理',
    requiresAuth: true,
    icon: 'icon-swap',
    order: 4,
  },
  children: [
    {
      path: 'spot',
      name: 'SpotTransaction',
      component: () => import('@/views/admin/transaction/spot-list.vue'),
      meta: {
        locale: '现货交易',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'lever',
      name: 'LeverTransaction',
      component: () => import('@/views/admin/transaction/lever-list.vue'),
      meta: {
        locale: '永续合约',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'lever-multiple',
      name: 'LeverMultiple',
      component: () => import('@/views/admin/transaction/lever-multiple.vue'),
      meta: {
        locale: '永续合约倍数管理',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'micro',
      name: 'MicroTransaction',
      component: () => import('@/views/admin/transaction/micro-list.vue'),
      meta: {
        locale: '交割合约',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'micro-config',
      name: 'MicroConfig',
      component: () => import('@/views/admin/transaction/micro-config.vue'),
      meta: {
        locale: '交割合约周期配置',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default ADMIN_TRANSACTION;
