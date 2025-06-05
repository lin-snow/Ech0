declare namespace App {
  /**
   * Namespace Api
   */
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
      data: T
    }

    namespace Auth {
      type LoginParams = {
        username: string
        password: string
      }

      type LoginResponse = {
        token: string
      }

      type SignupParams = {
        username: string
        password: string
      }
    }

    namespace User {
      type User = {
        id: number
        username: string
        password?: string
        is_admin: boolean
        avatar?: string
      }

      type UserInfo = {
        username: string
        password: string
        is_admin: boolean
        avatar: string
      }

      type UserStatus = {
        user_id: number
        username: string
        is_admin: boolean
      }
    }

    namespace Ech0 {
      type ParamsByPagination = {
        page: number
        pageSize: number
        search?: string
      }

      type Echo = {
        id: number
        content: string
        username: string
        image_url: string
        image_source: string
        images: Image[]
        private: boolean
        user_id: number
        extension: string
        extension_type: string
        created_at: string
      }

      type Image = {
        id: number
        message_id: number
        image_url: string
        image_source: string
      }

      type ImageToAdd = {
        image_url: string
        image_source: string
      }

      type EchoToAdd = {
        content: string
        image_url?: string | null
        image_source?: string | null
        images?: ImageToAdd[] | null
        extension?: string | null
        extension_type?: string | null
        private: boolean
      }

      type PaginationResult = {
        items: Echo[]
        total: number
      }

      type Status = {
        sys_admin_id: number
        username: string
        logo: string
        users: App.Api.User.UserStatus[]
        total_messages: number
      }

      type HeatMap = {
        date: string
        count: number
      }[]

      type ImageToDelete = {
        url: string
        source: string
      }

      type GithubCardData = {
        name: string
        stargazers_count: number
        forks_count: number
        description: string
        owner: {
          avatar_url: string
        }
      }
    }

    namespace Setting {
      type SystemSetting = {
        site_title: string
        server_name: string
        server_url: string
        allow_register: boolean
        ICP_number: string
        meting_api: string
        custom_css: string
        custom_js: string
      }
    }

    namespace Connect {
      type Connect = {
        server_name: string
        server_url: string
        logo: string
        ech0s: number
        sys_username: string
      }

      type Connected = {
        id: number
        connect_url: string
      }
    }

    namespace Todo {
      type Todo = {
        id: number
        content: string
        user_id: number
        username: string
        status: number
        created_at: string
      }

      type TodoToAdd = {
        content: string
      }
    }
  }
}
