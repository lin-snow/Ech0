<template>
  <UCard class="mx-auto sm:max-w-sm hover:shadow-md">
    <div class="flex justify-between mb-3">
      <div class="flex justify-start items-center gap-2">
        <UIcon name="i-fluent-emoji-flat-alien-monster" class="w-6 h-6" />
        <h2 class="text-lg font-bold italic text-slate-600">Ech0s~</h2>
      </div>
      <div class="flex gap-2">
        <ClientOnly>
          <a href="/rss" target="_blank">
            <UIcon name="i-mdi-rss" class="w-5 h-5 text-gray-400" title="RSS" />
          </a>
        </ClientOnly>
        <NuxtLink to="https://github.com/lin-snow/Ech0/" target="_blank">
          <UIcon name="i-mdi-github" class="w-5 h-5 text-gray-400" title="GitHub" />
        </NuxtLink>
        <NuxtLink to="/status">
          <UIcon name="i-mdi-server-outline" class="w-5 h-5 text-gray-400" title="Status" />
        </NuxtLink>
      </div>
    </div>

    <div>
      <VditorEditor ref="vditorEditor" v-model="MessageContent" />
      <div class="flex justify-between items-center">
        <div class="flex items-center justify-start gap-2">
          <!-- <UInput size="sm" color="gray" :trailing="true" placeholder="你の名字" v-model="Username" class="w-24" /> -->
          <input
            id="file-input"
            ref="fileInput"
            type="file"
            accept="image/*"
            @change="addImage"
            class="hidden"
            placeholder="选择图片"
          />
          <!-- 使用 NuxtUI 的 UButton 触发文件选择 -->
          <UButton
            color="gray"
            variant="solid"
            class="cursor-pointer"
            size="sm"
            icon="i-fluent-image-20-regular"
            @click="triggerFileInput"
            title="上传图片"
          />
          <!-- 是否设为私密？ -->
          <UButton
            color="gray"
            variant="solid"
            size="sm"
            @click="Private = !Private"
            :icon="privateIcon"
            title="设为私密？"
           />
        </div>

        <div class="flex gap-2">
          <!-- 清空表单 -->
          <UButton
            icon="i-fluent-broom-16-regular"
            variant="solid"
            color="gray"
            size="sm"
            @click="clearForm"
            title="清空内容"
          />

          <!-- 添加Ech0 -->
          <UButton
            icon="i-fluent-add-12-filled"
            variant="solid"
            color="gray"
            size="sm"
            @click="addMessage"
            title="发布Ech0"
          />
        </div>
      </div>
    </div>

    <!-- 显示图片 -->
    <div
      v-if="ImageUrl"
      class="w-5/6 h-auto shadow-md rounded-md mx-auto mt-5 overflow-hidden"
    >
      <a :href="`${BASE_API}${ImageUrl}`" data-fancybox>
        <img :src="`${BASE_API}${ImageUrl}`" alt="" loading="lazy" />
      </a>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import type { MessageToSave } from "~/types/models";
import { UButton } from "#components";
import { useMessage } from "~/composables/useMessage";
import { Fancybox } from "@fancyapps/ui";
import "@fancyapps/ui/dist/fancybox/fancybox.css";
import VditorEditor from './VditorEditor.vue'

const BASE_API = useRuntimeConfig().public.baseApi;
const { save, uploadImage } = useMessage();

const Username = ref("");
const MessageContent = ref("");
const Private = ref<boolean>(false);
const ImageUrl = ref("");
const fileInput = ref<HTMLInputElement | null>(null);
const vditorEditor = ref<InstanceType<typeof VditorEditor> | null>(null); // 添加 vditorEditor 引用

const privateIcon = computed(() => (Private.value ? 'i-mdi-eye-off-outline' : 'i-mdi-eye-outline'));

const clearForm = () => {
  Username.value = "";
  MessageContent.value = "";
  ImageUrl.value = "";
  Private.value = false;
  // 清空 Vditor 编辑器内容
  if (vditorEditor.value) {
    vditorEditor.value.clear();
  }
};

const addMessage = async () => {
  const message: MessageToSave = {
    username: Username.value,
    content: MessageContent.value,
    private: Private.value,
    image_url: ImageUrl.value,
  };

  const response = await save(message);
  if (response) {
    clearForm(); // 清空表单
  }
};

const triggerFileInput = () => {
  // 手动触发 file input 的点击事件
  const input = document.getElementById("file-input");
  if (input) {
    input.click();
  }
};

const addImage = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files ? input.files[0] : null;

  if (!file) {
    console.error("没有选择文件");
    return;
  }

  const imageUrl = await uploadImage(file);

  if (!imageUrl) {
    console.error("上传图片失败");
    return;
  }

  ImageUrl.value = imageUrl; // 设置上传后的图片URL
};

onMounted(() => {
  Fancybox.bind("[data-fancybox]", {});
});

onBeforeUnmount(() => {
  Fancybox.destroy();
});
</script>
