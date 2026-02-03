import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_CURRENCY: AppRouteRecordRaw = {
  path: '/admin/currency',
  name: 'CurrencyManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '币种管理',
    requiresAuth: true,
    icon: 'icon-apps',
    order: 2,
  },
  children: [
    {
      path: 'list',
      name: 'CurrencyList',
      component: () => import('@/views/admin/currency/list.vue'),
      meta: {
        locale: '币种列表',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default ADMIN_CURRENCY;
