import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_WALLET: AppRouteRecordRaw = {
  path: '/admin/wallet',
  name: 'WalletManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '财务管理',
    requiresAuth: true,
    icon: 'icon-storage',
    order: 3,
  },
  children: [
    {
      path: 'deposit-review',
      name: 'DepositReview',
      component: () => import('@/views/admin/wallet/deposit-review.vue'),
      meta: {
        locale: '充值审核',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'withdrawal',
      name: 'WithdrawalList',
      component: () => import('@/views/admin/wallet/withdrawal-list.vue'),
      meta: {
        locale: '提现审核',
        requiresAuth: true,
        roles: ['*'],
      },
    },
  ],
};

export default ADMIN_WALLET;
