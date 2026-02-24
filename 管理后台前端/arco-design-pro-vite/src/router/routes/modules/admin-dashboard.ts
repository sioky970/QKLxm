import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_DASHBOARD: AppRouteRecordRaw = {
  path: '/admin',
  name: 'Admin',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '仪表盘',
    requiresAuth: true,
    icon: 'icon-dashboard',
    order: 0,
  },
  children: [
    {
      path: 'dashboard',
      name: 'Dashboard',
      component: () => import('@/views/admin/dashboard/index.vue'),
      meta: {
        locale: '仪表盘',
        requiresAuth: true,
        roles: ['*'],
        icon: 'icon-dashboard',
      },
    },
  ],
};

export default ADMIN_DASHBOARD;
