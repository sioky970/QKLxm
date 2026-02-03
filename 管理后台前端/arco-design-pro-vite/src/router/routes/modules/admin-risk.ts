import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_RISK: AppRouteRecordRaw = {
  path: '/admin/risk',
  name: 'RiskManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '风控管理',
    requiresAuth: true,
    icon: 'icon-safe',
    order: 5,
  },
  children: [
    {
      path: 'config',
      name: 'RiskConfig',
      component: () => import('@/views/admin/risk/config.vue'),
      meta: {
        locale: '全局风控',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default ADMIN_RISK;
