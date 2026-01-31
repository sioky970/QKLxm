import { DirectiveBinding } from 'vue';
import { useUserStore } from '@/store';

/**
 * 权限指令
 * 用法：v-permission="['user:add', 'user:edit']"
 *
 * 如果用户没有指定权限，则移除该元素
 */
function checkPermission(el: HTMLElement, binding: DirectiveBinding) {
  const { value } = binding;
  const userStore = useUserStore();

  if (value && value instanceof Array && value.length > 0) {
    const permissionRoles = value;
    const { role } = userStore;

    // 超级管理员拥有所有权限
    if (role === 'admin') {
      return true;
    }

    // 检查是否有匹配的权限
    const hasPermission =
      permissionRoles.includes('*') || permissionRoles.includes(role);

    if (!hasPermission && el.parentNode) {
      el.parentNode.removeChild(el);
    }

    return hasPermission;
  }

  return true;
}

export default {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    checkPermission(el, binding);
  },
  updated(el: HTMLElement, binding: DirectiveBinding) {
    checkPermission(el, binding);
  },
};
