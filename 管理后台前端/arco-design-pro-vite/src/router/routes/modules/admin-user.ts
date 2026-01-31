import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ADMIN_USER: AppRouteRecordRaw = {
  path: '/admin/user',
  name: 'UserManagement',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: '用户管理',
    requiresAuth: true,
    icon: 'icon-user',
    order: 1,
  },
  children: [
    {
      path: 'list',
      name: 'UserList',
      component: () => import('@/views/admin/user/list.vue'),
      meta: {
        locale: '用户列表',
        requiresAuth: true,
        roles: ['*'],
      },
    },
    {
      path: 'detail/:id',
      name: 'UserDetail',
      component: () => import('@/views/admin/user/detail.vue'),
      meta: {
        locale: '用户详情',
        requiresAuth: true,
        roles: ['*'],
        hideInMenu: true,
      },
    },
  ],
};

export default ADMIN_USER;
