// 处理LocalStorage的工具函数

export const localStg = {
  /**
   * setItem
   * @param key
   * @param obj
   */
  setItem<T>(key: string, obj: T) {
    localStorage.setItem(key, JSON.stringify(obj))
  },

  /**
   * getItem
   * @param key
   * @returns
   */
  getItem<T>(key: string): T | null {
    const item = localStorage.getItem(key)
    if (!item) return null
    try {
      return JSON.parse(item) as T
    } catch {
      return null
    }
  },

  /**
   * removeItem
   * @param key
   */
  removeItem(key: string) {
    localStorage.removeItem(key)
  },

  /**
   * clear
   */
  clear() {
    localStorage.clear()
  },
}
