#!/bin/bash

RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
ORANGE='\033[38;5;214m'
RESET='\033[0m'
if [ -n "${NO_COLOR:-}" ] || [ "${TERM:-}" = "dumb" ]; then
  RED=''
  GREEN=''
  YELLOW=''
  ORANGE=''
  RESET=''
fi

REPO="${ECH0_GITHUB_REPO:-lin-snow/Ech0}"
BRANCH="${ECH0_GITHUB_BRANCH:-main}"
SCRIPT_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}/scripts/ech0.sh"
SERVICE_NAME="ech0"
UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
INSTALL_PATH_DEFAULT="/opt/ech0"
MANAGER_PATH="/usr/local/sbin/em"
COMMAND_LINK="/usr/local/bin/em"
COMMAND_LINK_COMPAT="/usr/local/bin/ech0"
LEGACY_MANAGER_PATH="/usr/local/sbin/ech0-manager"
DEFAULT_PORT="6277"

ARCH=""
LATEST_TAG=""
FORCE_UPDATE="false"
WORK_DIR=""
DOWNLOAD_FILE=""

handle_error() {
  local code=$1
  shift
  echo -e "${RED}错误: $*${RESET}"
  exit "$code"
}

log_info() {
  echo -e "${GREEN}$*${RESET}"
}

log_warn() {
  echo -e "${YELLOW}$*${RESET}"
}

log_error() {
  echo -e "${RED}$*${RESET}"
}

ensure_command() {
  command -v "$1" >/dev/null 2>&1 || handle_error 1 "缺少依赖: $1"
}

ensure_root() {
  if [ "$(id -u)" -ne 0 ]; then
    handle_error 1 "请使用 root 权限运行此命令"
  fi
}

detect_arch() {
  local machine
  machine=$(uname -m)

  case "$machine" in
    x86_64|amd64)
      ARCH="amd64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      handle_error 1 "当前架构(${machine})不支持，仅支持 amd64 与 arm64"
      ;;
  esac
}

resolve_install_path() {
  local target="${1:-$INSTALL_PATH_DEFAULT}"
  target="${target%/}"
  if [[ $target != */ech0 ]]; then
    target="$target/ech0"
  fi
  echo "$target"
}

get_installed_path() {
  if [ -f "$UNIT_FILE" ]; then
    local service_path
    service_path=$(grep -m1 '^WorkingDirectory=' "$UNIT_FILE" | cut -d'=' -f2-)
    if [ -n "$service_path" ] && [ -f "$service_path/ech0" ]; then
      echo "$service_path"
      return 0
    fi
  fi

  if [ -f "$INSTALL_PATH_DEFAULT/ech0" ]; then
    echo "$INSTALL_PATH_DEFAULT"
    return 0
  fi

  echo "$INSTALL_PATH_DEFAULT"
  return 1
}

is_ech0_installed() {
  local installed_path
  installed_path=$(get_installed_path)
  [ -f "$installed_path/ech0" ]
}

init_work_dir() {
  if [ -z "$WORK_DIR" ]; then
    WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/ech0-install.XXXXXX") \
      || handle_error 1 "无法创建临时目录"
    DOWNLOAD_FILE="$WORK_DIR/ech0.tar.gz"
  fi
}

cleanup_tmp() {
  if [ -n "$WORK_DIR" ] && [ -d "$WORK_DIR" ]; then
    rm -rf "$WORK_DIR"
  fi
}

trap cleanup_tmp EXIT

github_curl() {
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -fsSL --connect-timeout 10 --max-time 60 \
      -H "Authorization: Bearer ${GITHUB_TOKEN}" "$@"
  else
    curl -fsSL --connect-timeout 10 --max-time 60 "$@"
  fi
}

resolve_latest_tag() {
  if [ -n "$LATEST_TAG" ]; then
    return
  fi

  local tag
  tag=$(github_curl "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
    | grep -oE '"tag_name"[[:space:]]*:[[:space:]]*"v[^"]+"' \
    | head -n1 \
    | sed -E 's/.*"(v[^"]+)".*/\1/')

  # 回退：`/releases/latest` 不可用时，从列表中取最新的正式版本（排除 rc/beta 等预发布）
  if [ -z "$tag" ]; then
    tag=$(github_curl "https://api.github.com/repos/${REPO}/releases?per_page=30" 2>/dev/null \
      | grep -oE '"tag_name"[[:space:]]*:[[:space:]]*"v[0-9]+\.[0-9]+\.[0-9]+"' \
      | head -n1 \
      | sed -E 's/.*"(v[^"]+)".*/\1/')
  fi

  if [ -z "$tag" ]; then
    handle_error 1 "无法解析 Ech0 最新版本（GitHub API 可能被限流，可设置 GITHUB_TOKEN 后重试）"
  fi

  LATEST_TAG="$tag"
}

installed_version() {
  local path=$1
  if [ ! -x "$path/ech0" ]; then
    return 1
  fi

  local version
  version=$(cd "$path" && ./ech0 version 2>/dev/null \
    | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' \
    | head -n1)
  if [ -z "$version" ]; then
    return 1
  fi

  echo "$version"
}

download_package() {
  detect_arch
  resolve_latest_tag
  init_work_dir

  local asset="ech0-linux-${ARCH}.tar.gz"
  local url="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${asset}"

  log_info "目标版本: ${LATEST_TAG}"
  log_info "下载 ${asset}"

  if ! curl -fL --proto '=https' --retry 3 --retry-delay 3 \
    --connect-timeout 10 --max-time 600 "$url" -o "$DOWNLOAD_FILE"; then
    handle_error 1 "下载 Ech0 失败，请检查网络连接"
  fi
}

prepare_install_dir() {
  local install_path=$1
  local parent
  parent=$(dirname "$install_path")

  mkdir -p "$parent" || handle_error 1 "无法创建目录: $parent"

  if [ -f "$install_path/ech0" ]; then
    handle_error 1 "目标位置已存在 Ech0，请使用 update"
  fi

  mkdir -p "$install_path" || handle_error 1 "无法创建安装目录: $install_path"
}

extract_release_package_dir() {
  init_work_dir
  local extract_dir="$WORK_DIR/extract"
  rm -rf "$extract_dir"
  mkdir -p "$extract_dir" || handle_error 1 "无法创建解压目录"

  if ! tar -xzf "$DOWNLOAD_FILE" -C "$extract_dir"; then
    handle_error 1 "解压失败"
  fi

  local binary_path
  binary_path=$(find "$extract_dir" -mindepth 1 -maxdepth 5 -type f \
    \( -name "ech0" -o -name "ech0-linux-${ARCH}" \) -print -quit)
  if [ -z "$binary_path" ]; then
    handle_error 1 "未找到 Ech0 二进制文件"
  fi

  dirname "$binary_path"
}

# 把发行包内容安装到目标目录：数据目录与 .env 永远不会被覆盖
install_package_files() {
  local package_dir=$1
  local target_path=$2

  shopt -s dotglob nullglob
  for item in "$package_dir"/*; do
    local name
    name=$(basename "$item")
    case "$name" in
      data|.env)
        continue
        ;;
      "ech0-linux-${ARCH}")
        install -m 0755 "$item" "$target_path/ech0" || return 1
        ;;
      *)
        rm -rf "${target_path:?}/$name"
        cp -a "$item" "$target_path/" || return 1
        ;;
    esac
  done
  shopt -u dotglob nullglob

  [ -f "$target_path/ech0" ] || return 1
  chmod 755 "$target_path/ech0" || return 1
}

ensure_service() {
  local install_path=$1
  local force=${2:-false}

  if [ -f "$UNIT_FILE" ] && [ "$force" != "true" ]; then
    return
  fi

  cat <<EOF >"$UNIT_FILE"
[Unit]
Description=Ech0 Service
Documentation=https://github.com/${REPO}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=${install_path}
ExecStart=${install_path}/ech0 serve
Restart=on-failure
RestartSec=3
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME" >/dev/null 2>&1
}

wait_active() {
  local timeout=${1:-20}
  local i=0

  while [ "$i" -lt "$timeout" ]; do
    if systemctl is-active --quiet "$SERVICE_NAME"; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done

  return 1
}

config_value() {
  local path=$1
  local key=$2
  local value=""

  value=$(printenv "$key" 2>/dev/null || true)
  if [ -z "$value" ] && [ -f "$path/.env" ]; then
    value=$(grep -E "^[[:space:]]*${key}[[:space:]]*=" "$path/.env" | tail -n1 \
      | sed -E 's/^[^=]*=[[:space:]]*//' \
      | sed -E "s/^[\"']//; s/[\"'][[:space:]]*\$//")
  fi

  echo "$value"
}

health_check() {
  local install_path=$1
  local host port
  host=$(config_value "$install_path" ECH0_SERVER_HOST)
  port=$(config_value "$install_path" ECH0_SERVER_PORT)

  case "$host" in
    ""|0.0.0.0|::)
      host="127.0.0.1"
      ;;
  esac
  [ -n "$port" ] || port="$DEFAULT_PORT"

  curl -fsS --connect-timeout 2 --max-time 5 "http://${host}:${port}/healthz"
}

running_version() {
  local install_path=$1
  local body
  body=$(health_check "$install_path" 2>/dev/null) || return 1
  echo "$body" \
    | grep -oE '"version"[[:space:]]*:[[:space:]]*"[^"]+"' \
    | head -n1 \
    | sed -E 's/.*"([^"]+)"$/\1/'
}

verify_service() {
  local install_path=$1

  if ! wait_active 20; then
    log_error "Ech0 服务未能在预期时间内启动"
    journalctl -u "$SERVICE_NAME" -n 20 --no-pager 2>/dev/null || true
    return 1
  fi

  local version
  if version=$(running_version "$install_path"); then
    log_info "健康检查通过 (运行版本 v${version})"
  else
    log_warn "服务进程已启动，但健康检查未通过（若自定义了监听地址/端口可忽略）"
  fi

  return 0
}

resolve_self_path() {
  local path="$0"
  local resolved
  local base

  if command -v readlink >/dev/null 2>&1; then
    resolved=$(readlink -f "$path" 2>/dev/null) || resolved=""
    if [ -n "$resolved" ]; then
      path="$resolved"
    fi
  fi

  [ -f "$path" ] || return 1
  base=$(basename "$path")
  case "$base" in
    bash|sh|dash|zsh|ksh|busybox)
      return 1
      ;;
  esac
  grep -q '^SERVICE_NAME="ech0"' "$path" 2>/dev/null || return 1

  echo "$path"
}

install_cli_helper() {
  local source_path
  source_path=$(resolve_self_path)

  mkdir -p "$(dirname "$MANAGER_PATH")"

  if [ -n "$source_path" ]; then
    cp "$source_path" "$MANAGER_PATH" || log_warn "复制管理脚本失败"
  else
    # 通过管道运行（curl ... | bash）时没有脚本本体，回退到远程下载
    local tmp_file
    tmp_file=$(mktemp) || log_warn "无法创建临时文件"
    if [ -z "$tmp_file" ]; then
      return 1
    fi
    if fetch_script "$tmp_file"; then
      mv -f "$tmp_file" "$MANAGER_PATH" || log_warn "写入管理脚本失败"
    else
      rm -f "$tmp_file"
      log_warn "下载管理脚本失败，跳过管理命令安装"
      return 1
    fi
  fi

  chmod 755 "$MANAGER_PATH" 2>/dev/null || true
  link_manager_command
}

fetch_script() {
  local dest=$1
  curl -fsSL --proto '=https' --connect-timeout 10 --max-time 60 \
    "$SCRIPT_URL" -o "$dest" || return 1
  grep -q '^SERVICE_NAME="ech0"' "$dest" 2>/dev/null || return 1
  chmod 755 "$dest" 2>/dev/null || true
}

link_manager_command() {
  mkdir -p "$(dirname "$COMMAND_LINK")"
  if [ "$COMMAND_LINK" != "$MANAGER_PATH" ]; then
    ln -sf "$MANAGER_PATH" "$COMMAND_LINK" || log_warn "创建命令链接失败"
  fi
  if [ "$COMMAND_LINK_COMPAT" != "$MANAGER_PATH" ] && [ "$COMMAND_LINK_COMPAT" != "$COMMAND_LINK" ]; then
    ln -sf "$MANAGER_PATH" "$COMMAND_LINK_COMPAT" || log_warn "创建命令链接失败"
  fi
}

# 从远端脚本仓库刷新已安装的管理脚本；allow_create=true 时允许新建（self-update）
refresh_manager() {
  local allow_create=${1:-false}
  local dir
  local tmp_file

  if [ ! -f "$MANAGER_PATH" ] && [ "$allow_create" != "true" ]; then
    return 0
  fi

  dir=$(dirname "$MANAGER_PATH")
  mkdir -p "$dir" || {
    log_warn "无法创建目录: ${dir}"
    return 1
  }

  tmp_file=$(mktemp "${dir}/.em.XXXXXX") || {
    log_warn "无法创建临时文件，跳过管理脚本更新"
    return 1
  }

  if ! fetch_script "$tmp_file"; then
    rm -f "$tmp_file"
    log_warn "获取管理脚本失败（${SCRIPT_URL}），跳过管理脚本更新"
    return 1
  fi

  if [ -f "$MANAGER_PATH" ] && cmp -s "$tmp_file" "$MANAGER_PATH"; then
    rm -f "$tmp_file"
    log_info "管理脚本已是最新: ${MANAGER_PATH}"
    return 0
  fi

  if ! mv -f "$tmp_file" "$MANAGER_PATH"; then
    rm -f "$tmp_file"
    log_warn "替换管理脚本失败: ${MANAGER_PATH}"
    return 1
  fi
  chmod 755 "$MANAGER_PATH" 2>/dev/null || true
  link_manager_command

  log_info "管理脚本已更新: ${MANAGER_PATH}"
  return 0
}

self_update() {
  ensure_root
  ensure_command curl
  refresh_manager true
}

should_install_cli_helper() {
  if [ ! -t 0 ]; then
    return 0
  fi

  local choice
  read -r -p "是否安装管理命令(em，并兼容 ech0)? [Y/n]: " choice
  case "$choice" in
    n|N)
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

install_ech0() {
  ensure_root
  ensure_command curl
  ensure_command tar
  ensure_command systemctl

  local installed_path
  installed_path=$(get_installed_path)
  if [ -f "$installed_path/ech0" ]; then
    handle_error 1 "检测到 Ech0 已安装在 ${installed_path}，请使用 update 或先 uninstall（仅升级管理脚本可用 self-update）"
  fi

  local target_path
  target_path=$(resolve_install_path "$1")

  if [ -f "$target_path/ech0" ]; then
    handle_error 1 "目标位置已存在 Ech0 (${target_path})，请使用 update 或先 uninstall"
  fi

  download_package
  local package_dir
  package_dir=$(extract_release_package_dir)

  prepare_install_dir "$target_path"
  install_package_files "$package_dir" "$target_path" || handle_error 1 "写入程序文件失败"
  ensure_service "$target_path" true

  systemctl restart "$SERVICE_NAME"
  if ! verify_service "$target_path"; then
    handle_error 1 "Ech0 启动失败，请查看 systemctl status ${SERVICE_NAME} 与日志"
  fi

  local cli_installed="false"
  if should_install_cli_helper; then
    if install_cli_helper; then
      cli_installed="true"
    fi
  else
    log_warn "已跳过管理命令安装，可稍后重新运行本脚本来安装"
  fi

  log_info "Ech0 安装完成"
  echo "安装路径: $target_path"
  echo "服务命令: systemctl {start|stop|restart|status} ${SERVICE_NAME}"
  if [ "$cli_installed" = "true" ]; then
    echo "管理命令: ${COMMAND_LINK} (兼容命令: ${COMMAND_LINK_COMPAT})"
    echo "查看命令: ${COMMAND_LINK} help"
  fi
}

update_ech0() {
  ensure_root
  ensure_command curl
  ensure_command tar
  ensure_command systemctl

  local target_path
  target_path=$(get_installed_path)

  if [ ! -f "$target_path/ech0" ]; then
    handle_error 1 "未检测到已安装的 Ech0"
  fi

  detect_arch
  resolve_latest_tag
  refresh_manager

  local current_version
  current_version=$(installed_version "$target_path" 2>/dev/null || true)

  if [ "$FORCE_UPDATE" != "true" ] && [ -n "$current_version" ] && [ "$current_version" = "$LATEST_TAG" ]; then
    log_info "当前已是最新版本 ${current_version}（如需强制重装请使用 update --force）"
    return
  fi

  if [ -n "$current_version" ]; then
    log_info "当前版本: ${current_version}"
  else
    log_warn "无法识别当前版本，将直接覆盖安装"
  fi

  download_package
  local package_dir
  package_dir=$(extract_release_package_dir)

  local backup="$target_path/ech0.bak"
  systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true
  cp -p "$target_path/ech0" "$backup" || handle_error 1 "备份旧版本失败"

  if ! install_package_files "$package_dir" "$target_path"; then
    mv -f "$backup" "$target_path/ech0" 2>/dev/null || true
    handle_error 1 "更新文件失败，已恢复旧版本"
  fi

  ensure_service "$target_path"
  systemctl restart "$SERVICE_NAME"

  if ! verify_service "$target_path"; then
    log_warn "新版本启动失败，正在回滚到 ${current_version:-旧版本}"
    systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true
    if mv -f "$backup" "$target_path/ech0"; then
      chmod 755 "$target_path/ech0"
      systemctl restart "$SERVICE_NAME"
      verify_service "$target_path" >/dev/null 2>&1 \
        || handle_error 1 "回滚后服务仍未启动，请查看 journalctl -u ${SERVICE_NAME}"
      handle_error 1 "已回滚到更新前版本，请查看 journalctl -u ${SERVICE_NAME} 排查失败原因"
    fi
    handle_error 1 "回滚失败，备份位于 ${backup}"
  fi

  rm -f "$backup"
  log_info "Ech0 更新完成: ${current_version:-unknown} -> ${LATEST_TAG}"
}

uninstall_ech0() {
  ensure_root
  ensure_command systemctl

  local target_path
  target_path=$(get_installed_path)

  if [ ! -f "$target_path/ech0" ] && [ ! -f "$UNIT_FILE" ]; then
    handle_error 1 "未检测到 Ech0 安装"
  fi

  read -r -p "确认卸载 Ech0? [y/N]: " choice
  case "$choice" in
    y|Y)
      local delete_data_choice
      local remove_data="false"
      local data_dir="$target_path/data"
      if [ -d "$data_dir" ]; then
        read -r -p "检测到数据目录 ${data_dir}，是否同时删除(删除后不可恢复)? [y/N]: " delete_data_choice
        case "$delete_data_choice" in
          y|Y)
            remove_data="true"
            ;;
        esac
      fi

      systemctl stop "$SERVICE_NAME" >/dev/null 2>&1 || true
      systemctl disable "$SERVICE_NAME" >/dev/null 2>&1 || true
      rm -f "$UNIT_FILE"
      systemctl daemon-reload

      if [ "$remove_data" = "true" ]; then
        rm -rf "$target_path"
        log_info "已删除程序与数据目录"
      else
        shopt -s dotglob nullglob
        for item in "$target_path"/*; do
          local name
          name=$(basename "$item")
          case "$name" in
            data|.env)
              continue
              ;;
          esac
          rm -rf "$item"
        done
        shopt -u dotglob nullglob
        log_warn "已保留数据目录与配置: ${data_dir}, ${target_path}/.env"
      fi

      rm -f "$MANAGER_PATH" "$COMMAND_LINK" "$COMMAND_LINK_COMPAT" "$LEGACY_MANAGER_PATH"

      log_info "Ech0 已卸载"
      ;;
    *)
      log_warn "已取消卸载"
      ;;
  esac
}

show_status() {
  ensure_command systemctl

  local target_path
  target_path=$(get_installed_path)

  if [ -f "$target_path/ech0" ]; then
    local version
    version=$(installed_version "$target_path" 2>/dev/null || true)
    echo "安装路径: ${target_path}"
    echo "安装版本: ${version:-未知}"
  else
    log_warn "未检测到 Ech0 安装"
  fi

  if systemctl is-active --quiet "$SERVICE_NAME"; then
    log_info "Ech0 服务运行中"
    local version
    if version=$(running_version "$target_path"); then
      echo "运行版本: v${version}"
    fi
    echo "开机自启: $(systemctl is-enabled "$SERVICE_NAME" 2>/dev/null || echo unknown)"
  else
    log_error "Ech0 服务未运行"
    if [ -f "$UNIT_FILE" ]; then
      echo -e "最近日志: journalctl -u ${SERVICE_NAME} -n 50"
    fi
  fi
}

show_logs() {
  ensure_command journalctl

  case "${1:-}" in
    -f|--follow)
      journalctl -u "$SERVICE_NAME" -n 100 -f
      ;;
    *)
      journalctl -u "$SERVICE_NAME" -n 100 --no-pager
      ;;
  esac
}

render_script_header() {
  local command_name="${1:-menu}"
  local install_state="未安装"
  local install_state_display=""
  local install_path="-"
  local installed_version=""
  local service_state="未知"
  local service_state_display=""
  local separator="${ORANGE}==================================================${RESET}"

  if is_ech0_installed; then
    install_path=$(get_installed_path)
    install_state="已安装"
    installed_version=$(installed_version "$install_path" 2>/dev/null || true)
  fi

  if command -v systemctl >/dev/null 2>&1; then
    if [ -f "$UNIT_FILE" ]; then
      if systemctl is-active --quiet "$SERVICE_NAME"; then
        service_state="运行中"
      else
        service_state="未运行"
      fi
    else
      service_state="未安装服务"
    fi
  else
    service_state="systemctl 不可用"
  fi

  if [ "$install_state" = "已安装" ]; then
    install_state_display="${GREEN}${install_state}${RESET}"
  else
    install_state_display="${ORANGE}${install_state}${RESET}"
  fi

  case "$service_state" in
    运行中)
      service_state_display="${GREEN}${service_state}${RESET}"
      ;;
    未运行|未安装服务|systemctl\ 不可用)
      service_state_display="${ORANGE}${service_state}${RESET}"
      ;;
    *)
      service_state_display="${RED}${service_state}${RESET}"
      ;;
  esac

  echo -e "$separator"
  echo -e "${ORANGE}  _____     _      ___${RESET}"
  echo -e "${ORANGE} | ____|___| |__  / _ \\\\${RESET}"
  echo -e "${ORANGE} |  _| / __| '_ \\\\| | | |${RESET}"
  echo -e "${ORANGE} | |__| (__| | | | |_| |${RESET}"
  echo -e "${ORANGE} |_____|\\___|_| |_|\\___/${RESET}"
  echo -e "${ORANGE}Ech0 管理脚本 (安装/更新/卸载/服务管理)${RESET}"
  echo "当前命令: ${command_name}"
  echo -e "安装状态: ${install_state_display}"
  echo "安装路径: ${install_path}"
  if [ -n "$installed_version" ]; then
    echo "安装版本: ${installed_version}"
  fi
  echo -e "服务状态: ${service_state_display}"
  echo -e "$separator"
}

run_cli_command() {
  local target_path
  target_path=$(get_installed_path)
  if [ ! -f "$target_path/ech0" ]; then
    handle_error 1 "未检测到 Ech0 安装"
  fi

  (cd "$target_path" && "$target_path/ech0" "$@")
}

start_service() {
  ensure_root
  ensure_command systemctl
  systemctl start "$SERVICE_NAME"
  log_info "Ech0 已启动"
}

stop_service() {
  ensure_root
  ensure_command systemctl
  systemctl stop "$SERVICE_NAME"
  log_info "Ech0 已停止"
}

restart_service() {
  ensure_root
  ensure_command systemctl
  systemctl restart "$SERVICE_NAME"
  log_info "Ech0 已重启"
}

show_menu() {
  local install_label="1) 安装 Ech0"
  if is_ech0_installed; then
    install_label="1) 安装 Ech0 (已安装)"
  fi

  echo
  echo "欢迎使用 Ech0 部署脚本"
  echo "$install_label"
  echo "2) 更新 Ech0"
  echo "3) 删除 Ech0"
  echo "4) 查看服务状态"
  echo "5) 查看版本信息"
  echo "6) 启动服务"
  echo "7) 停止服务"
  echo "8) 重启服务"
  echo "9) 查看服务日志"
  echo "10) 更新管理脚本"
  echo "0) 退出脚本"
  echo
  read -r -p "请选择 [0-9/10]: " choice

  case "$choice" in
    1)
      if is_ech0_installed; then
        local installed_path
        installed_path=$(get_installed_path)
        log_warn "已检测到安装: ${installed_path}，请使用更新或先卸载"
        return
      fi
      read -r -p "安装路径(默认: ${INSTALL_PATH_DEFAULT}): " path_choice
      install_ech0 "$path_choice"
      ;;
    2)
      update_ech0
      ;;
    3)
      uninstall_ech0
      ;;
    4)
      show_status
      ;;
    5)
      run_cli_command version
      ;;
    6)
      start_service
      ;;
    7)
      stop_service
      ;;
    8)
      restart_service
      ;;
    9)
      show_logs
      ;;
    10)
      self_update
      ;;
    0)
      exit 0
      ;;
    *)
      log_error "无效的选项"
      ;;
  esac
}

usage() {
  cat <<EOF
用法: $0 [命令]
  install [路径]   安装 Ech0 (默认安装到 ${INSTALL_PATH_DEFAULT})
  update [--force] 更新到最新版本（--force 可重装同版本）
  uninstall        删除 Ech0 及 systemd 服务
  status           查看服务状态
  version          查看已安装版本
  logs [-f]        查看服务日志（-f 持续跟踪）
  self-update      更新管理脚本自身（update 也会自动刷新）
  start            启动服务
  stop             停止服务
  restart          重启服务
  tui              运行 ech0 tui
  help             显示本帮助
  (无参数)         进入交互菜单

环境变量:
  ECH0_GITHUB_REPO   GitHub 仓库，默认 ${REPO}
  GITHUB_TOKEN       GitHub API 限流时可用
EOF
}

main() {
  render_script_header "$1"

  case "$1" in
    install)
      install_ech0 "$2"
      ;;
    update)
      shift
      for arg in "$@"; do
        case "$arg" in
          -f|--force)
            FORCE_UPDATE="true"
            ;;
        esac
      done
      update_ech0
      ;;
    uninstall)
      uninstall_ech0
      ;;
    status)
      show_status
      ;;
    version)
      run_cli_command version
      ;;
    logs)
      show_logs "${2:-}"
      ;;
    self-update)
      self_update
      ;;
    start)
      start_service
      ;;
    tui)
      run_cli_command tui
      ;;
    stop)
      stop_service
      ;;
    restart)
      restart_service
      ;;
    help|-h|--help)
      usage
      ;;
    "")
      while true; do
        show_menu
      done
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
