import axios from 'axios';
import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import { Message, Modal } from '@arco-design/web-vue';
import { useUserStore } from '@/store';
import { getToken } from '@/utils/auth';

export interface HttpResponse<T = unknown> {
  type: string;
  message: string;
  error: string;
  data: T;
}

if (import.meta.env.VITE_API_BASE_URL) {
  axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;
}

axios.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // let each request carry token
    // this example using the JWT token
    // Authorization is a custom headers key
    // please modify it according to the actual situation
    const token = getToken();
    if (token) {
      if (!config.headers) {
        config.headers = {};
      }
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    // do something
    return Promise.reject(error);
  }
);
// add response interceptors
axios.interceptors.response.use(
  (response: AxiosResponse<HttpResponse>) => {
    const res = response.data;

    // 打印响应详情到控制台
    console.log('=== API Response ===');
    console.log('URL:', response.config.url);
    console.log('Status:', response.status);
    console.log('Response Data:', res);
    console.log('==================');

    // 兼容两种响应格式：
    // 1. 标准格式: {type: 'success'/'ok', message, data}
    // 2. LayUI表格格式: {code: 0, msg, count, data}
    const isSuccess = 
      res.type === 'success' || 
      res.type === 'ok' || 
      (res as any).code === 0;

    if (!isSuccess) {
      const errorMsg = res.error || res.message || (res as any).msg || 'Error';
      console.error('API Error:', errorMsg);

      Message.error({
        content: errorMsg,
        duration: 5 * 1000,
      });

      // 401 未授权
      if (response.status === 401) {
        Modal.error({
          title: '确认退出',
          content: '您已退出登录，可以取消停留在此页面，或重新登录',
          okText: '重新登录',
          async onOk() {
            const userStore = useUserStore();
            await userStore.logout();
            window.location.reload();
          },
        });
      }
      return Promise.reject(new Error(errorMsg));
    }
    return response;
  },
  (error: any) => {
    console.error('=== API Request Error ===');
    console.error('Error:', error);
    console.error('Message:', error.message);
    if (error.response) {
      console.error('Response Status:', error.response.status);
      console.error('Response Data:', error.response.data);
    }
    console.error('========================');

    Message.error({
      content: error.message || 'Request Error',
      duration: 5 * 1000,
    });
    return Promise.reject(error);
  }
);
