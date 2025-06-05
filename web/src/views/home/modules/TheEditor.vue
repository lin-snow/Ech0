<template>
  <div
    class="bg-white rounded-lg ring-1 ring-gray-200 ring-inset mx-auto shadow-sm hover:shadow-md"
  >
    <div class="mx-auto w-full px-3 py-4">
      <!-- Title && Nav -->
      <div class="flex justify-between items-center py-1 px-3">
        <div class="flex flex-row items-center gap-2 justify-between">
          <!-- <div class="text-xl">👾</div> -->
          <div>
            <img
              :src="logo"
              alt="logo"
              class="w-6 sm:w-7 h-6 sm:h-7 rounded-full ring-1 ring-gray-200 shadow-sm object-cover"
            />
          </div>
          <h1 class="text-slate-600 font-bold italic sm:text-xl">
            {{ SystemSetting.server_name }}
          </h1>
        </div>
        <div class="flex flex-row items-center gap-2">
          <!-- Github -->
          <div>
            <a href="https://github.com/lin-snow/Ech0" target="_blank" title="Github">
              <Github class="w-6 sm:w-7 h-6 sm:h-7 text-gray-400" />
            </a>
          </div>
        </div>
      </div>

      <!-- Editor -->
      <div class="rounded-lg p-2 sm:p-3 mb-1">
        <TheMdEditor
          v-model="echoToAdd.content"
          class="rounded-lg"
          v-if="currentMode === Mode.ECH0"
        />

        <!-- ImageMode -->
        <div v-if="currentMode === Mode.Image">
          <h2 class="text-gray-500 font-bold">插入图片（支持本地上传、直链）</h2>
          <p class="text-gray-400 text-sm mb-2">注意：仅允许添加一张</p>
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2">
              <span class="text-gray-500">选择添加方式：</span>
              <!-- 直链 -->
              <BaseButton
                :icon="Url"
                class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
                @click="imageToAdd.image_source = ImageSource.URL"
                title="插入图片链接"
              />
              <!-- 上传本地 -->
              <BaseButton
                :icon="Upload"
                class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
                @click="imageToAdd.image_source = ImageSource.LOCAL"
                title="上传本地图片"
              />
            </div>
            <div>
              <BaseButton
                v-if="imageToAdd.image_url != ''"
                :icon="Addmore"
                class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
                @click="handleAddMoreImage"
                title="添加更多图片"
              />
            </div>
          </div>
          <div class="my-1">
            <!-- 图片上传本地 -->
            <input
              id="file-input"
              class="hidden"
              type="file"
              accept="image/*"
              ref="fileInput"
              @change="handleUploadImage"
            />
            <BaseButton
              v-if="imageToAdd.image_source === ImageSource.LOCAL"
              @click="handleTriggerUpload"
              class="rounded-md"
              title="上传本地图片"
            >
              <span class="text-gray-400">点击上传</span>
            </BaseButton>
            <!-- 图片直链 -->
            <BaseInput
              v-if="imageToAdd.image_source === ImageSource.URL"
              v-model="imageToAdd.image_url"
              class="rounded-lg h-auto w-full"
              placeholder="请输入图片链接..."
            />
          </div>
        </div>

        <!-- todoMode -->
        <BaseTextArea
          v-if="currentMode === Mode.TODO"
          v-model="todoToAdd.content"
          class="rounded-lg h-auto sm:min-h-[6rem] md:min-h-[9rem]"
          placeholder="请输入待办事项..."
          :rows="3"
        />
        <!-- PlayMusic(上传音乐文件) -->
        <div v-if="currentMode === Mode.PlayMusic">
          <h2 class="text-gray-500 font-bold mb-1">欢迎使用音乐播放模式（仅PC）</h2>
          <div class="mb-1 flex items-center gap-2">
            <p class="text-gray-500">上传音乐：</p>
            <input
              id="file-input"
              class="hidden"
              type="file"
              accept="audio/*"
              ref="fileInput"
              @change="handleUploadMusic"
            />
            <BaseButton
              :icon="Audio"
              @click="handleTriggerUpload"
              class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
              title="上传音乐"
            />
          </div>
          <div class="flex items-center gap-2">
            <p class="text-gray-500">删除音乐：</p>
            <BaseButton
              :icon="Delete"
              @click="handleDeleteMusic"
              class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
              title="删除音乐"
            />
          </div>
        </div>

        <!-- Panel -->
        <TheModePanel
          v-if="currentMode === Mode.Panel"
          @switch-todo="handleSwitchTodoMode"
          @switch-extension="handleSwitchExtensionMode"
          @switch-play-music="handleSwitchPlayMusicMode"
        />
        <!-- Extension -->
        <div v-if="currentMode === Mode.EXTEN">
          <div v-if="currentExtensionType === ExtensionType.MUSIC">
            <h2 class="text-gray-500 font-bold mb-1">音乐分享（支持网易云/QQ音乐）</h2>
            <p class="text-gray-400 text-sm mb-1">注意：不支持VIP歌曲，建议使用自建API</p>
            <BaseInput
              v-model="extensionToAdd.extension"
              class="rounded-lg h-auto w-full"
              placeholder="音乐链接..."
            />
            <div
              v-if="
                extensionToAdd.extension.length > 0 &&
                extensionToAdd.extension_type === ExtensionType.MUSIC
              "
              class="mt-1 text-gray-400 text-md"
            >
              解析结果：
              <span v-if="parseMusicURL(extensionToAdd.extension)" class="text-green-400"
                >成功</span
              >
              <span v-else class="text-red-300">失败</span>
            </div>
          </div>
          <div v-if="currentExtensionType === ExtensionType.VIDEO">
            <div class="text-gray-500 font-bold mb-1">Bilibili视频分享(粘贴自动提取BV号)</div>
            <BaseInput
              v-model="bilibiliURL"
              class="rounded-lg h-auto w-full my-2"
              placeholder="B站分享链接..."
            />
            <div class="text-gray-500 my-1">提取的BV号为：{{ extensionToAdd.extension }}</div>
          </div>
          <div v-if="currentExtensionType === ExtensionType.GITHUBPROJ">
            <div class="text-gray-500 font-bold mb-1">Github项目分享</div>
            <BaseInput
              v-model="extensionToAdd.extension"
              class="rounded-lg h-auto w-full"
              placeholder="https://github.com/username/repo"
            />
          </div>
          <div v-if="currentExtensionType === ExtensionType.WEBSITE">
            <div class="text-gray-500 font-bold mb-1">网站链接分享</div>
            <!-- 网站标题 -->
            <BaseInput
              v-model="websiteToAdd.title"
              class="rounded-lg h-auto w-full mb-2"
              placeholder="网站标题..."
            />
            <BaseInput
              v-model="websiteToAdd.site"
              class="rounded-lg h-auto w-full"
              placeholder="https://example.com"
            />
          </div>
        </div>
      </div>

      <!-- Buttons -->
      <div class="flex flex-row items-center justify-between px-2">
        <div class="flex flex-row items-center gap-2">
          <!-- ShowMore -->
          <div>
            <BaseButton
              :icon="Advance"
              @click="handleChangeMode"
              :class="
                [
                  'w-8 h-8 sm:w-9 sm:h-9 rounded-md',
                  todoMode
                    ? 'bg-orange-100 shadow-[0_0_12px_-4px_rgba(255,140,0,0.6)] !ring-0 !text-white'
                    : '',
                ].join(' ')
              "
              title="其它"
            />
          </div>
          <!-- Photo Upload -->
          <div v-if="currentMode === Mode.ECH0">
            <BaseButton
              :icon="ImageUpload"
              @click="handleAddImageMode"
              class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
              title="添加图片"
            />
          </div>
          <!-- Privacy Set -->
          <div v-if="currentMode === Mode.ECH0">
            <BaseButton
              :icon="echoToAdd.private ? Private : Public"
              @click="handlePrivate"
              class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
              title="是否私密"
            />
          </div>
        </div>

        <div class="flex flex-row items-center gap-2">
          <!-- Clear -->
          <!-- <div>
            <BaseButton
              :icon="Clear"
              @click="handleClear"
              class="w-8 h-8 rounded-md"
              title="清空输入和图片"
            />
          </div> -->
          <!-- Publish -->
          <div v-if="currentMode !== Mode.Panel">
            <BaseButton
              :icon="Publish"
              @click="handleAdd"
              class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
              title="发布"
            />
          </div>
        </div>
      </div>

      <!-- Preview Image -->
      <div
        v-if="
          imagesToAdd &&
          imagesToAdd.length > 0 &&
          (currentMode === Mode.ECH0 || currentMode === Mode.Image)
        "
        class="relative rounded-lg shadow-lg w-5/6 mx-auto my-7"
      >
        <button
          @click="handleRemoveImage"
          class="absolute -top-3 -right-4 bg-red-100 hover:bg-red-300 text-gray-600 rounded-lg w-7 h-7 flex items-center justify-center shadow"
          title="移除图片"
        >
          <Close class="w-4 h-4" />
        </button>
        <div class="rounded-lg overflow-hidden">
          <a :href="getImageToAddUrl(imagesToAdd[imageIndex])" data-fancybox>
            <img
              :src="getImageToAddUrl(imagesToAdd[imageIndex])"
              alt="Image"
              class="max-w-full object-cover"
              loading="lazy"
            />
          </a>
        </div>
      </div>
      <!-- 图片切换 -->
      <div v-if="imagesToAdd.length > 1" class="flex items-center justify-center">
        <button @click="imageIndex = Math.max(imageIndex - 1, 0)">
          <Prev class="w-7 h-7" />
        </button>
        <span class="text-gray-500 text-sm mx-2">
          {{ imageIndex + 1 }} / {{ imagesToAdd.length }}
        </span>
        <button @click="imageIndex = Math.min(imageIndex + 1, imagesToAdd.length - 1)">
          <Next class="w-7 h-7" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Next from '@/components/icons/next.vue'
import Prev from '@/components/icons/prev.vue'
import Github from '@/components/icons/github.vue'
import Advance from '@/components/icons/advance.vue'
import Upload from '@/components/icons/upload.vue'
import Url from '@/components/icons/url.vue'
import Close from '@/components/icons/close.vue'
import Audio from '@/components/icons/audio.vue'
import ImageUpload from '@/components/icons/image.vue'
import Public from '@/components/icons/public.vue'
import Private from '@/components/icons/private.vue'
import Publish from '@/components/icons/publish.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheMdEditor from '@/components/advanced/TheMdEditor.vue'
import TheModePanel from './TheModePanel.vue'
import { theToast } from '@/utils/toast'
import { Fancybox } from '@fancyapps/ui'
import { onMounted, ref, watch } from 'vue'
import {
  fetchUploadImage,
  fetchAddEcho,
  fetchGetStatus,
  fetchAddTodo,
  fetchDeleteImage,
  fetchUploadMusic,
  fetchDeleteMusic,
} from '@/service/api'
import { useEchoStore } from '@/stores/echo'
import { useSettingStore } from '@/stores/settting'
import { useTodoStore } from '@/stores/todo'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { storeToRefs } from 'pinia'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import Delete from '@/components/icons/delete.vue'
import { parseMusicURL } from '@/utils/other'
import { getImageToAddUrl } from '@/utils/other'
import { getApiUrl } from '@/service/request/shared'
import Addmore from '@/components/icons/addmore.vue'

const emit = defineEmits(['refreshAudio'])

const echoStore = useEchoStore()
const todoStore = useTodoStore()
const settingStore = useSettingStore()

const { setTodoMode, getTodos } = todoStore
const { SystemSetting } = storeToRefs(settingStore)
const { todoMode } = storeToRefs(todoStore)

const enum Mode {
  ECH0 = 0,
  Panel = 1,
  TODO = 2,
  EXTEN = 3,
  PlayMusic = 4,
  Image = 5,
}
const enum ExtensionType {
  MUSIC = 'MUSIC',
  VIDEO = 'VIDEO',
  GITHUBPROJ = 'GITHUBPROJ',
  WEBSITE = 'WEBSITE',
}
const enum ImageSource {
  LOCAL = 'local',
  URL = 'url',
  S3 = 's3',
  R2 = 'r2',
}
const currentMode = ref<Mode>(Mode.ECH0)
const currentExtensionType = ref<ExtensionType>()

const apiUrl = getApiUrl()
const logo = ref<string>('/favicon.svg')

const handleChangeMode = () => {
  if (currentMode.value === Mode.ECH0) {
    currentMode.value = Mode.Panel
  } else if (
    currentMode.value === Mode.TODO ||
    currentMode.value === Mode.PlayMusic ||
    currentMode.value === Mode.Image
  ) {
    currentMode.value = Mode.ECH0
    setTodoMode(false)

    if (!echoToAdd.value.image_url || echoToAdd.value.image_url.length === 0) {
      echoToAdd.value.image_source = null
    }
  } else {
    currentMode.value = Mode.ECH0
  }
}
const handleSwitchExtensionMode = (extensiontype: ExtensionType) => {
  currentMode.value = Mode.EXTEN
  currentExtensionType.value = extensiontype
  extensionToAdd.value.extension_type = extensiontype
}
const handleSwitchTodoMode = () => {
  setTodoMode(true)
  currentMode.value = Mode.TODO
}
const handleSwitchPlayMusicMode = () => {
  currentMode.value = Mode.PlayMusic
}

const websiteToAdd = ref<{
  title: string
  site: string
}>({
  title: '',
  site: '',
})
const bilibiliURL = ref<string>('')
const extensionToAdd = ref({
  extension: '',
  extension_type: '',
})
const imageIndex = ref<number>(0)
const imageSourceMemory = ref<string>()
const imageToAdd = ref<App.Api.Ech0.ImageToAdd>({
  image_url: '',
  image_source: '',
})
const imagesToAdd = ref<App.Api.Ech0.ImageToAdd[]>([])
const echoToAdd = ref<App.Api.Ech0.EchoToAdd>({
  content: '',
  image_url: null,
  image_source: null,
  images: [],
  private: false,
  extension: null,
  extension_type: null,
})

const todoToAdd = ref<App.Api.Todo.TodoToAdd>({
  content: '',
})

const handleAddMoreImage = () => {
  imagesToAdd.value.push({
    image_url: String(imageToAdd.value.image_url),
    image_source: String(imageToAdd.value.image_source),
  })

  imageToAdd.value.image_url = ''
  imageToAdd.value.image_source = ''
}

const handleAddImageMode = () => {
  currentMode.value = Mode.Image
}
const fileInput = ref<HTMLInputElement | null>(null)
const handleTriggerUpload = () => {
  imageSourceMemory.value = imageToAdd.value.image_source
  if (fileInput.value) {
    fileInput.value.click()
  }
}
const handleUploadImage = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    fetchUploadImage(file).then((res) => {
      if (res.code === 1) {
        // 改成新版的图片数组
        // echoToAdd.value.image_url = res.data
        imageToAdd.value.image_url = String(res.data)
        imageToAdd.value.image_source = ImageSource.LOCAL
        theToast.success('图片上传成功！')

        handleAddMoreImage()

        // if (currentMode.value === Mode.Image) {
        //   currentMode.value = Mode.ECH0
        // }
      }
    })
  }
}

const handleUploadMusic = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    fetchUploadMusic(file).then((res) => {
      if (res.code === 1) {
        theToast.success('音乐上传成功！')
        emit('refreshAudio')
      }
    })
  }
}
const handleDeleteMusic = () => {
  if (confirm('确定要删除音乐吗？')) {
    fetchDeleteMusic().then((res) => {
      if (res.code === 1) {
        theToast.success('音乐删除成功！')
        emit('refreshAudio')
      }
    })
  }
}

const handleRemoveImage = () => {
  if (confirm('确定要移除图片吗？')) {
    const imageToDel: App.Api.Ech0.ImageToDelete = {
      url: String(imagesToAdd.value[imageIndex.value].image_url),
      source: String(imagesToAdd.value[imageIndex.value].image_source),
    }

    if (imageToDel.source === ImageSource.LOCAL) {
      fetchDeleteImage({
        url: imageToDel.url,
        source: imageToDel.source,
      }).then((res) => {
        if (res.code === 1) {
          // 从数组中删除图片
          imagesToAdd.value.splice(imageIndex.value, 1)
        }
      })
    } else {
      imagesToAdd.value.splice(imageIndex.value, 1)
    }

    imageIndex.value = 0
  }
}

const handlePrivate = () => {
  echoToAdd.value.private = !echoToAdd.value.private
}

const handleClear = () => {
  echoToAdd.value.content = ''
  echoToAdd.value.image_url = null
  echoToAdd.value.image_source = null
  echoToAdd.value.images = []
  echoToAdd.value.private = false
  echoToAdd.value.extension = null
  echoToAdd.value.extension_type = null
  extensionToAdd.value.extension = ''
  extensionToAdd.value.extension_type = ''
  bilibiliURL.value = ''
  imagesToAdd.value = []
  imageToAdd.value.image_url = ''
  imageToAdd.value.image_source = ''
}

const handleAddEcho = () => {
  echoToAdd.value.images = imagesToAdd.value

  // 检查是否有外部链接分享
  if (extensionToAdd.value.extension_type === ExtensionType.WEBSITE) {
    // 检查是否存在网站链接
    if (websiteToAdd.value.title.length > 0 && websiteToAdd.value.site.length === 0) {
      theToast.error('网站链接不能为空！')
      return
    }

    // 检查是否存在网站标题
    if (websiteToAdd.value.title.length === 0 && websiteToAdd.value.site.length > 0) {
      websiteToAdd.value.title = '外部链接'
    }

    if (websiteToAdd.value.title.length > 0 && websiteToAdd.value.site.length > 0) {
      // 将网站标题和链接添加到扩展中 (序列化为json)
      extensionToAdd.value.extension = JSON.stringify({
        title: websiteToAdd.value.title,
        site: websiteToAdd.value.site,
      })
    } else {
      extensionToAdd.value.extension = ''
      extensionToAdd.value.extension_type = ''
    }
  }

  if (extensionToAdd.value.extension.length > 0 && extensionToAdd.value.extension_type.length > 0) {
    echoToAdd.value.extension = extensionToAdd.value.extension
    echoToAdd.value.extension_type = extensionToAdd.value.extension_type
  } else {
    echoToAdd.value.extension = null
    echoToAdd.value.extension_type = null
  }

  if (
    !echoToAdd.value.content &&
    (!echoToAdd.value.images || echoToAdd.value.images.length === 0) &&
    !echoToAdd.value.extension &&
    !echoToAdd.value.extension_type
  ) {
    theToast.error('内容不能为空！')
    return
  }

  // if (!echoToAdd.value.image_url || echoToAdd.value.image_url.length === 0) {
  //   echoToAdd.value.image_source = null
  // }

  fetchAddEcho(echoToAdd.value).then((res) => {
    if (res.code === 1) {
      theToast.success('发布成功！')
      handleClear()
      echoStore.refreshEchos()
      currentMode.value = Mode.ECH0
    }
  })
}

const handleAddTodo = () => {
  if (todoToAdd.value.content === '') {
    theToast.error('内容不能为空！')
    return
  }

  fetchAddTodo(todoToAdd.value).then((res) => {
    if (res.code === 1) {
      theToast.success('添加成功！')
      todoToAdd.value.content = ''
      getTodos()
    }
  })
}

const handleAdd = () => {
  if (todoMode.value) {
    handleAddTodo()
  } else {
    handleAddEcho()
  }
}

// 监听用户输入
watch(
  () => bilibiliURL.value,
  (newVal) => {
    if (newVal.length > 0) {
      const bvRegex = /(BV[0-9A-Za-z]{10})/
      const match = newVal.match(bvRegex)
      if (match) {
        extensionToAdd.value.extension = match[0]
      } else {
        theToast.error('请输入正确的B站分享链接！')
      }
    }
  },
)

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
  fetchGetStatus().then((res) => {
    if (res.code === 1) {
      const theLogo = res.data.logo
      if (theLogo && theLogo !== '') {
        logo.value = `${apiUrl}${theLogo}`
      }
    }
  })
})
</script>
