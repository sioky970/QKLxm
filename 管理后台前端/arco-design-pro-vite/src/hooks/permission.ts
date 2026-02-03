/* eslint-disable */
import { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store';

export default function usePermission() {
  const userStore = useUserStore();

  return {
    /**
     * 检查用户是否有访问指定路由的权限
     */
    accessRouter(route: RouteLocationNormalized | RouteRecordRaw) {
      const { role } = userStore;
      
      // 不需要认证的路由直接放行
      if (!route.meta?.requiresAuth) {
        return true;
      }

      // 超级管理员拥有所有权限
      if (role === 'admin') {
        return true;
      }

      // 检查路由是否配置了角色限制
      const roles = route.meta?.roles as string[] | undefined;
      
      if (!roles || roles.length === 0) {
        return true;
      }

      // 检查用户角色是否在允许的角色列表中
      return roles.includes('*') || roles.includes(role || '');
    },

    /**
     * 查找用户有权限访问的第一个路由
     */
    findFirstPermissionRoute(
      routes: RouteRecordRaw[],
      role = 'user'
    ): RouteRecordRaw | null {
      const cloneRoutes = [...routes];
      
      while (cloneRoutes.length) {
        const route = cloneRoutes.shift();
        
        if (!route) continue;

        const roles = route.meta?.roles as string[] | undefined;
        
        // 如果没有角色限制或用户角色在允许列表中
        if (!roles || roles.length === 0 || roles.includes('*') || roles.includes(role)) {
          // 如果有子路由，继续查找
          if (route.children && route.children.length > 0) {
            const firstChild = this.findFirstPermissionRoute(route.children, role);
            if (firstChild) return firstChild;
          }
          
          // 返回当前路由
          return route;
        }

        // 如果有子路由，将子路由加入队列继续查找
        if (route.children && route.children.length > 0) {
          cloneRoutes.push(...route.children);
        }
      }

      return null;
    },

    /**
     * 检查用户是否有指定权限（用于按钮级别的权限控制）
     */
    hasPermission(permission: string | string[]): boolean {
      const { role } = userStore;

      // 超级管理员拥有所有权限
      if (role === 'admin') {
        return true;
      }

      if (typeof permission === 'string') {
        return role === permission || permission === '*';
      }

      if (Array.isArray(permission)) {
        return permission.includes('*') || permission.includes(role || '');
      }

      return false;
    },
  };
}
