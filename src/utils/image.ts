/**
 * 图片URL处理工具
 * 用于处理后端返回的图片路径，确保在开发和生产环境下都能正确访问
 */

/**
 * 获取完整的图片URL
 * @param path - 后端返回的图片路径（如：/uploads/kyc/xxx.png）
 * @returns 完整的图片URL
 * 
 * 开发环境：直接使用相对路径，由 Vite 代理到后端
 * 生产环境：拼接 API 基础地址
 */
export function getImageUrl(path: string): string {
  if (!path) return '';
  
  // 如果已经是完整URL，直接返回
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  
  // 开发环境：直接使用相对路径，由 Vite 代理
  if (import.meta.env.DEV) {
    return path;
  }
  
  // 生产环境：拼接 API 基础地址
  const baseURL = import.meta.env.VITE_API_BASE_URL || '';
  return `${baseURL}${path}`;
}

/**
 * 获取上传文件的完整URL
 * @param path - 文件路径
 * @returns 完整的文件URL
 */
export function getUploadUrl(path: string): string {
  return getImageUrl(path);
}

/**
 * 检查图片URL是否有效
 * @param url - 图片URL
 * @returns Promise<boolean>
 */
export async function checkImageUrl(url: string): Promise<boolean> {
  if (!url) return false;
  
  try {
    const response = await fetch(url, { method: 'HEAD' });
    return response.ok;
  } catch (error) {
    console.error('检查图片URL失败:', error);
    return false;
  }
}

/**
 * 获取默认的占位图
 * @returns 占位图URL
 */
export function getPlaceholderImage(): string {
  return 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgZmlsbD0iI2Y1ZjVmNSIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LXNpemU9IjE4IiBmaWxsPSIjYzBjMGMwIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBkeT0iLjNlbSI+5peg5Zu+54mHPC90ZXh0Pjwvc3ZnPg==';
}
