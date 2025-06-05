// 处理全局通知工具函数
import { useToast, TYPE, POSITION } from 'vue-toastification'

const toast = useToast()
const toastDefaultOptions = {
  position: POSITION.TOP_RIGHT,
  timeout: 2000,
  draggable: false,
}

interface customToastOptions {
  timeout?: number
  pauseOnHover?: boolean
  closeOnClick?: boolean
  onClick?: () => void
  onClose?: () => void
}

export const theToast = {
  /**
   * 显示成功通知
   * @param content
   * @param options
   */
  success(content: string, options?: customToastOptions) {
    toast(content, {
      type: TYPE.SUCCESS,
      ...toastDefaultOptions,
      ...options,
    })
  },

  /**
   * 显示错误通知
   * @param content
   * @param options
   */
  error(content: string, options?: customToastOptions) {
    toast(content, {
      type: TYPE.ERROR,
      ...toastDefaultOptions,
      ...options,
    })
  },

  /**
   * 显示信息通知
   * @param content
   * @param options
   */
  info(content: string, options?: customToastOptions) {
    toast(content, {
      type: TYPE.INFO,
      ...toastDefaultOptions,
      ...options,
    })
  },

  /**
   * 显示警告通知
   * @param content
   * @param options
   */
  warning(content: string, options?: customToastOptions) {
    toast(content, {
      type: TYPE.WARNING,
      ...toastDefaultOptions,
      ...options,
    })
  },
}
