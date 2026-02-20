import { useThemeStore } from './theme'
import { useUserStore } from './user'
import { useSettingStore } from './setting'
import { useTodoStore } from './todo'
import { useEchoStore } from './echo'
import { useZoneStore } from './zone'
import { useEditorStore } from './editor'
import { useInboxStore } from './inbox'

export async function initStores() {
  const themeStore = useThemeStore()
  const userStore = useUserStore()
  const settingStore = useSettingStore()
  const todoStore = useTodoStore()
  const echoStore = useEchoStore()
  const zoneStore = useZoneStore()
  const editorStore = useEditorStore()
  const inboxStore = useInboxStore()

  themeStore.init()
  await userStore.init()
  await settingStore.init()
  todoStore.init()
  editorStore.init()
  echoStore.init()
  zoneStore.init()
  inboxStore.init()
}
