#!/bin/bash

###############################################################################
# Ech0 Manager Script
# Version: 1.0.0
# Last Updated: 2025-05-16
# Description:
#   A management script for Ech0.
#   Provides installation, update, uninstallation and management functions.
# Requirements:
#   - Linux with systemd
#   - Root privileges for installation, update, uninstallation
#   - curl, tar
#   - x86_64 architecture (amd64)
# Adapted from Alist Manager Script by Troray
###############################################################################

# 在脚本开头添加错误处理函数
handle_error() {
    local exit_code=$1
    local error_msg=$2
    echo -e "${RED_COLOR}错误：${error_msg}${RES}"
    exit ${exit_code}
}

# 配置部分
#######################
# GitHub 相关配置 - !!! YOU MUST CHANGE THESE !!!
GH_OWNER="lin-snow" # Your GitHub username or organization
GH_REPO="Ech0"  # Your GitHub repository name for ech0

GH_BASE_URL="https://github.com/${GH_OWNER}/${GH_REPO}/releases/latest/download"
APP_NAME="ech0"
TARBALL_NAME_PREFIX="${APP_NAME}-release-linux" # Used to build the full tarball name
BINARY_NAME="${APP_NAME}" # Assumed binary name inside the tarball
SERVICE_NAME="${APP_NAME}"
DEFAULT_PORT="6277" # !!! CHANGE IF ECH0 USES A DIFFERENT DEFAULT PORT !!!

# Installer script's own URL - IMPORTANT: Ensure this URL is correct and publicly accessible
SELF_INSTALLER_URL="http://echo.soopy.cn/install.sh"
#######################

# 颜色配置
RED_COLOR='\e[1;31m'
GREEN_COLOR='\e[1;32m'
YELLOW_COLOR='\e[1;33m'
RES='\e[0m'

# 管理脚本路径配置
MANAGER_PATH="/usr/local/sbin/${APP_NAME}-manager"  # 管理脚本存放路径
COMMAND_LINK="/usr/local/bin/${APP_NAME}"          # 命令软链接路径

# 获取已安装的 Ech0 路径
GET_INSTALLED_PATH() {
    # 从 service 文件中获取工作目录
    if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
        installed_path=$(grep "WorkingDirectory=" "/etc/systemd/system/${SERVICE_NAME}.service" | cut -d'=' -f2)
        if [ -f "$installed_path/${BINARY_NAME}" ]; then
            echo "$installed_path"
            return 0
        fi
    fi
    # 如果未找到或路径无效，返回默认路径
    echo "/opt/${APP_NAME}"
}

# 设置安装路径
if [ ! -n "$2" ]; then
    INSTALL_PATH="/opt/${APP_NAME}"
else
    INSTALL_PATH=${2%/}
    if ! [[ $INSTALL_PATH == */${APP_NAME} ]]; then
        INSTALL_PATH="$INSTALL_PATH/${APP_NAME}"
    fi
    # 创建父目录（如果不存在）
    parent_dir=$(dirname "$INSTALL_PATH")
    if [ ! -d "$parent_dir" ]; then
        mkdir -p "$parent_dir" || handle_error 1 "无法创建目录 $parent_dir"
    fi

    # 在创建目录后再检查权限
    if ! [ -w "$parent_dir" ]; then
        handle_error 1 "目录 $parent_dir 没有写入权限"
    fi
fi

# 如果是更新或卸载操作，使用已安装的路径
if [ "$1" = "update" ] || [ "$1" = "uninstall" ]; then
    INSTALL_PATH=$(GET_INSTALLED_PATH)
fi

clear

# 获取平台架构
if command -v arch >/dev/null 2>&1; then
    platform=$(arch)
else
    platform=$(uname -m)
fi

ARCH_SUFFIX=""
if [ "$platform" = "x86_64" ]; then
    ARCH_SUFFIX="amd64"
else
    echo -e "\r\n${RED_COLOR}错误${RES}，此脚本目前仅支持 x86_64 (amd64) 平台，与您的发布文件 ${TARBALL_NAME_PREFIX}-amd64.tar.gz 对应。\r\n"
    exit 1
fi

TARBALL_NAME="${TARBALL_NAME_PREFIX}-${ARCH_SUFFIX}.tar.gz"
GH_DOWNLOAD_URL="${GH_BASE_URL}" # Full URL will be constructed with TARBALL_NAME later

# 权限和环境检查
if [ "$(id -u)" != "0" ]; then
    if [ "$1" = "install" ] || [ "$1" = "update" ] || [ "$1" = "uninstall" ] || [ "$1" = "start" ] || [ "$1" = "stop" ] || [ "$1" = "restart" ]; then
        echo -e "\r\n${RED_COLOR}错误：请使用 root 权限运行此命令！${RES}\r\n"
        echo -e "提示：使用 ${GREEN_COLOR}sudo $0 $1${RES} 重试\r\n"
        exit 1
    fi
elif ! command -v systemctl >/dev/null 2>&1; then
    handle_error 1 "未找到 systemctl 命令。此脚本需要 systemd 环境。"
elif ! command -v curl >/dev/null 2>&1; then
    handle_error 1 "未找到 curl 命令，请先安装 curl。"
elif ! command -v tar >/dev/null 2>&1; then
    handle_error 1 "未找到 tar 命令，请先安装 tar。"
fi

CHECK() {
    # 检查目标目录是否存在，如果不存在则创建
    if [ ! -d "$(dirname "$INSTALL_PATH")" ]; then
        echo -e "${GREEN_COLOR}目录不存在，正在创建...${RES}"
        mkdir -p "$(dirname "$INSTALL_PATH")" || handle_error 1 "无法创建目录 $(dirname "$INSTALL_PATH")"
    fi
    # 检查是否已安装
    if [ -f "$INSTALL_PATH/${BINARY_NAME}" ] && [ "$1" = "install" ]; then # Only exit if trying to install over existing
        echo -e "${YELLOW_COLOR}警告：${APP_NAME} 已安装在 $INSTALL_PATH。${RES}"
        echo -e "如需重新安装，请先卸载或选择其他路径。"
        echo -e "如需更新，请使用 'update' 命令。"
        exit 0
    fi
    # 创建或清空安装目录 (only for new install)
    if [ "$1" = "install" ]; then
        if [ ! -d "$INSTALL_PATH/" ]; then
            mkdir -p "$INSTALL_PATH" || handle_error 1 "无法创建安装目录 $INSTALL_PATH"
        else
            # Optionally, ask before cleaning an existing directory if re-running install without uninstall
            # For now, let's assume if we reach here with 'install' and dir exists, it's okay to clear if binary isn't there.
            # Or, better, if binary ISN'T there, we clear. If it IS, the check above handles it.
            if [ ! -f "$INSTALL_PATH/${BINARY_NAME}" ]; then
                 echo -e "${YELLOW_COLOR}安装目录 $INSTALL_PATH 已存在但 ${BINARY_NAME} 未找到，将清空并重新使用该目录。${RES}"
                 rm -rf "${INSTALL_PATH:?}"/* # Protect against empty INSTALL_PATH
                 mkdir -p "$INSTALL_PATH" || handle_error 1 "无法重新创建安装目录 $INSTALL_PATH"
            fi
        fi
    fi
    echo -e "${GREEN_COLOR}安装目录准备就绪：$INSTALL_PATH${RES}"
}

# 添加下载函数，包含重试机制
download_file() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry_count=0
    local wait_time=5

    echo -e "${GREEN_COLOR}正在从 ${url} 下载到 ${output} ...${RES}"
    while [ $retry_count -lt $max_retries ]; do
        if curl -L --connect-timeout 10 --retry 3 --retry-delay 3 "$url" -o "$output"; then
            if [ -f "$output" ] && [ -s "$output" ]; then  # 检查文件是否存在且不为空
                return 0
            else
                echo -e "${YELLOW_COLOR}下载的文件为空或不存在，可能是下载链接错误或文件未成功保存。${RES}"
            fi
        fi
        
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo -e "${YELLOW_COLOR}下载失败，${wait_time} 秒后进行第 $((retry_count + 1)) 次重试...${RES}"
            sleep $wait_time
            wait_time=$((wait_time + 5))  # 每次重试增加等待时间
        else
            echo -e "${RED_COLOR}下载失败，已重试 $max_retries 次。请检查网络连接和URL: ${url}${RES}"
            return 1
        fi
    done
    return 1
}

INSTALL() {
    CURRENT_DIR=$(pwd) # 保存当前目录

    # 询问是否使用代理
    echo -e "${GREEN_COLOR}是否使用 GitHub 代理？（默认无代理）${RES}"
    echo -e "${GREEN_COLOR}代理地址必须为 https 开头，斜杠 / 结尾 ${RES}"
    echo -e "${GREEN_COLOR}例如：https://ghproxy.com/ ${RES}"
    read -p "请输入代理地址或直接按回车继续: " proxy_input

    local effective_download_url
    if [ -n "$proxy_input" ]; then
        effective_download_url="${proxy_input}${GH_DOWNLOAD_URL}/${TARBALL_NAME}"
        echo -e "${GREEN_COLOR}已使用代理地址: $proxy_input${RES}"
    else
        effective_download_url="${GH_DOWNLOAD_URL}/${TARBALL_NAME}"
        echo -e "${GREEN_COLOR}使用默认 GitHub 地址进行下载${RES}"
    fi

    echo -e "\r\n${GREEN_COLOR}下载 ${APP_NAME} (${TARBALL_NAME})...${RES}"
    if ! download_file "${effective_download_url}" "/tmp/${TARBALL_NAME}"; then
        handle_error 1 "下载 ${APP_NAME} 失败！"
    fi

    echo -e "${GREEN_COLOR}解压文件到 $INSTALL_PATH ...${RES}"
    # Ensure INSTALL_PATH exists before extracting
    mkdir -p "$INSTALL_PATH" || handle_error 1 "无法创建最终安装目录 $INSTALL_PATH"

    if ! tar zxf "/tmp/${TARBALL_NAME}" -C "$INSTALL_PATH/" --strip-components=1; then
        rm -f "/tmp/${TARBALL_NAME}"
        handle_error 1 "解压失败！请检查tarball是否有效、目标路径是否有写入权限或压缩包结构是否正确（是否需要 --strip-components=1）。"
    fi
    
    # Check if the binary exists after extraction
    # This assumes the binary is extracted directly into INSTALL_PATH.
    # If it's inside a subdirectory within the tarball, you need to add 'mv' commands here.
    # Example: if tarball extracts to $INSTALL_PATH/ech0-release-linux-amd64/ech0
    # then you would need:
    # if [ -d "$INSTALL_PATH/ech0-release-linux-amd64" ] && [ -f "$INSTALL_PATH/ech0-release-linux-amd64/${BINARY_NAME}" ]; then
    #   mv "$INSTALL_PATH/ech0-release-linux-amd64/${BINARY_NAME}" "$INSTALL_PATH/"
    #   rm -rf "$INSTALL_PATH/ech0-release-linux-amd64"
    # else
    #   # Handle case where expected subdirectory or binary within it is not found
    # fi

    if [ -f "$INSTALL_PATH/${BINARY_NAME}" ]; then
        echo -e "${GREEN_COLOR}下载和解压成功，正在安装...${RES}"
        chmod +x "$INSTALL_PATH/${BINARY_NAME}"
    else
        echo -e "${RED_COLOR}安装失败！${BINARY_NAME} 未在 $INSTALL_PATH 中找到。${RES}"
        echo -e "${YELLOW_COLOR}请检查 ${TARBALL_NAME} 的内容结构以及解压命令。${RES}"
        echo -e "${YELLOW_COLOR}如果压缩包内有顶层目录，脚本已尝试使用 --strip-components=1。如果仍然失败，请手动检查压缩包。${RES}"
        rm -f "/tmp/${TARBALL_NAME}"
        exit 1
    fi

    # 清理临时文件
    rm -f "/tmp/${TARBALL_NAME}"
    cd "$CURRENT_DIR" # 切回原目录
}

INIT() {
    if [ ! -f "$INSTALL_PATH/${BINARY_NAME}" ]; then
        handle_error 1 "当前系统未正确安装 ${APP_NAME} 或 ${BINARY_NAME} 文件丢失于 $INSTALL_PATH"
    fi

    echo -e "${GREEN_COLOR}创建 systemd 服务文件...${RES}"
    cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=${APP_NAME} service
Wants=network-online.target
After=network-online.target network.service

[Service]
Type=simple
WorkingDirectory=${INSTALL_PATH}
ExecStart=${INSTALL_PATH}/${BINARY_NAME} # !!! ADJUST 'server' IF YOUR APP STARTS DIFFERENTLY !!!
KillMode=process
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}" >/dev/null 2>&1 || echo -e "${YELLOW_COLOR}警告：无法自动启用 ${SERVICE_NAME} 服务。${RES}"
}

SUCCESS() {
    # clear # Usually cleared before menu, let's keep output visible
    print_line() {
        local text="$1"
        local width=51 # Adjusted for potentially shorter names
        printf "│ %-${width}s │\n" "$text"
    }

    # 获取本地 IP
    LOCAL_IP=$(ip addr show | grep -w inet | grep -v "127.0.0.1" | awk '{print $2}' | cut -d/ -f1 | head -n1)
    if [ -z "$LOCAL_IP" ]; then LOCAL_IP="无法获取"; fi

    # 获取公网 IP
    PUBLIC_IP=$(curl -s4 --connect-timeout 5 ip.sb || curl -s4 --connect-timeout 5 ifconfig.me || echo "获取失败")

    echo -e "┌────────────────────────────────────────────────────┐"
    print_line "${APP_NAME} 安装成功！"
    print_line ""
    print_line "访问地址 (默认端口 ${DEFAULT_PORT})："
    print_line "  局域网：http://${LOCAL_IP}:${DEFAULT_PORT}/"
    print_line "  公网：  http://${PUBLIC_IP}:${DEFAULT_PORT}/"
    # If Ech0 has a config file, you can add its path here:
    # print_line "配置文件：$INSTALL_PATH/data/config.json" # Example
    print_line ""
    echo -e "└────────────────────────────────────────────────────┘"

    # 安装命令行工具
    if ! INSTALL_CLI; then
        echo -e "${YELLOW_COLOR}警告：命令行工具安装失败，但不影响 ${APP_NAME} 的使用${RES}"
    fi

    echo -e "\n${GREEN_COLOR}启动服务中...${RES}"
    if systemctl restart "${SERVICE_NAME}"; then
        echo -e "${GREEN_COLOR}${SERVICE_NAME} 服务已启动。${RES}"
    else
        echo -e "${RED_COLOR}错误：${SERVICE_NAME} 服务启动失败。请检查日志: sudo journalctl -u ${SERVICE_NAME}${RES}"
    fi
    
    echo -e "管理: 在任意目录输入 ${GREEN_COLOR}${APP_NAME}${RES} 打开管理菜单"
    echo -e "\n${YELLOW_COLOR}温馨提示：如果端口无法访问，请检查服务器安全组、防火墙以及 ${APP_NAME} 服务状态和日志。${RES}"
    echo
    # exit 0 # Do not exit if called from menu loop
}

INSTALL_CLI() {
    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED_COLOR}错误：安装命令行工具需要 root 权限${RES}"
        return 1
    fi

    # Create directories for manager script and command link
    mkdir -p "$(dirname "$MANAGER_PATH")" || {
        echo -e "${RED_COLOR}错误：无法创建目录 $(dirname "$MANAGER_PATH")${RES}"
        return 1
    }
    chmod 755 "$(dirname "$MANAGER_PATH")"

    mkdir -p "$(dirname "$COMMAND_LINK")" || {
        echo -e "${RED_COLOR}错误：无法创建目录 $(dirname "$COMMAND_LINK")${RES}"
        # Attempt to clean up manager path dir if command link dir fails
        # This is a best-effort cleanup, might not always be necessary or perfect
        # Consider if rm -rf "$(dirname "$MANAGER_PATH")" is too aggressive if it contained other things
        return 1
    }

    echo -e "${GREEN_COLOR}正在从 ${SELF_INSTALLER_URL} 下载最新的管理脚本到 ${MANAGER_PATH} ...${RES}"
    if ! curl -sSL "${SELF_INSTALLER_URL}" -o "${MANAGER_PATH}"; then
        echo -e "${RED_COLOR}错误：无法下载管理脚本从 ${SELF_INSTALLER_URL} 到 ${MANAGER_PATH}${RES}"
        echo -e "${RED_COLOR}请检查网络连接以及URL是否正确: ${SELF_INSTALLER_URL}${RES}"
        # Clean up manager path if download fails, to avoid leaving a potentially empty/corrupt file
        rm -f "${MANAGER_PATH}" 
        return 1
    fi
    
    chmod 755 "$MANAGER_PATH" || {
        echo -e "${RED_COLOR}错误：设置 ${MANAGER_PATH} 权限失败${RES}"
        rm -f "$MANAGER_PATH" # Clean up downloaded script if chmod fails
        return 1
    }

    ln -sf "$MANAGER_PATH" "$COMMAND_LINK" || {
        echo -e "${RED_COLOR}错误：创建命令链接 $COMMAND_LINK 失败${RES}"
        rm -f "$MANAGER_PATH" # Clean up manager script if symlink fails
        return 1
    }

    echo -e "${GREEN_COLOR}命令行工具已成功安装/更新！${RES}"
    echo -e "现在你可以使用以下命令："
    echo -e "1. ${GREEN_COLOR}${APP_NAME}${RES}          - 快捷命令 (打开管理菜单)"
    echo -e "2. ${GREEN_COLOR}${APP_NAME}-manager${RES}  - 完整命令 (例如: ${APP_NAME}-manager install)"
    return 0
}

UPDATE() {
    if [ ! -f "$INSTALL_PATH/${BINARY_NAME}" ]; then
        handle_error 1 "未在 $INSTALL_PATH 找到 ${APP_NAME}，无法更新。"
    fi
    echo -e "${GREEN_COLOR}开始更新 ${APP_NAME} ...${RES}"

    # 询问是否使用代理 (与安装时逻辑相同)
    echo -e "${GREEN_COLOR}是否使用 GitHub 代理？（默认无代理）${RES}"
    read -p "请输入代理地址或直接按回车继续: " proxy_input

    local effective_download_url
    if [ -n "$proxy_input" ]; then
        effective_download_url="${proxy_input}${GH_DOWNLOAD_URL}/${TARBALL_NAME}"
        echo -e "${GREEN_COLOR}已使用代理地址: $proxy_input${RES}"
    else
        effective_download_url="${GH_DOWNLOAD_URL}/${TARBALL_NAME}"
        echo -e "${GREEN_COLOR}使用默认 GitHub 地址进行下载${RES}"
    fi

    echo -e "${GREEN_COLOR}停止 ${APP_NAME} 服务...${RES}"
    systemctl stop "${SERVICE_NAME}"

    echo -e "${GREEN_COLOR}备份当前 ${BINARY_NAME} 到 /tmp/${BINARY_NAME}.bak ...${RES}"
    if ! cp "$INSTALL_PATH/${BINARY_NAME}" "/tmp/${BINARY_NAME}.bak"; then
        echo -e "${RED_COLOR}备份旧版 ${BINARY_NAME} 失败，更新终止。${RES}"
        systemctl start "${SERVICE_NAME}" # 尝试恢复服务
        exit 1
    fi
    
    echo -e "${GREEN_COLOR}下载新版本 ${APP_NAME} (${TARBALL_NAME})...${RES}"
    if ! download_file "${effective_download_url}" "/tmp/${TARBALL_NAME}"; then
        echo -e "${RED_COLOR}下载失败，更新终止。正在恢复之前的版本...${RES}"
        mv "/tmp/${BINARY_NAME}.bak" "$INSTALL_PATH/${BINARY_NAME}" # 恢复备份
        systemctl start "${SERVICE_NAME}"
        exit 1
    fi

    # 创建临时解压目录
    UPDATE_TEMP_DIR="/tmp/${APP_NAME}_update_$$"
    mkdir -p "$UPDATE_TEMP_DIR" || {
        echo -e "${RED_COLOR}创建临时解压目录 ${UPDATE_TEMP_DIR} 失败，更新终止。${RES}"
        mv "/tmp/${BINARY_NAME}.bak" "$INSTALL_PATH/${BINARY_NAME}" # 恢复备份
        systemctl start "${SERVICE_NAME}"
        rm -f "/tmp/${TARBALL_NAME}" # 清理下载的压缩包
        exit 1
    }

    echo -e "${GREEN_COLOR}解压新版本到临时目录 ${UPDATE_TEMP_DIR} ...${RES}"
    if ! tar zxf "/tmp/${TARBALL_NAME}" -C "$UPDATE_TEMP_DIR/" --strip-components=1; then
        echo -e "${RED_COLOR}解压失败，更新终止。正在恢复之前的版本...${RES}"
        mv "/tmp/${BINARY_NAME}.bak" "$INSTALL_PATH/${BINARY_NAME}" # 恢复备份
        systemctl start "${SERVICE_NAME}"
        rm -f "/tmp/${TARBALL_NAME}"
        rm -rf "$UPDATE_TEMP_DIR" # 清理临时解压目录
        exit 1
    fi

    # 定位新版二进制文件 (根据你提供的结构，它直接在解压后的根目录)
    NEW_BINARY_PATH="$UPDATE_TEMP_DIR/${BINARY_NAME}" 

    if [ ! -f "$NEW_BINARY_PATH" ]; then
        echo -e "${RED_COLOR}错误：在新版本解压文件中未找到预期的二进制文件 ${NEW_BINARY_PATH}。${RES}"
        echo -e "${GREEN_COLOR}正在恢复之前的版本...${RES}"
        mv "/tmp/${BINARY_NAME}.bak" "$INSTALL_PATH/${BINARY_NAME}" # 恢复备份
        systemctl start "${SERVICE_NAME}"
        rm -f "/tmp/${TARBALL_NAME}"
        rm -rf "$UPDATE_TEMP_DIR"
        exit 1
    fi

    echo -e "${GREEN_COLOR}复制新版 ${BINARY_NAME} 到 ${INSTALL_PATH} ...${RES}"
    if ! cp "$NEW_BINARY_PATH" "$INSTALL_PATH/${BINARY_NAME}"; then
        echo -e "${RED_COLOR}复制新版 ${BINARY_NAME} 失败，更新终止。正在恢复之前的版本...${RES}"
        mv "/tmp/${BINARY_NAME}.bak" "$INSTALL_PATH/${BINARY_NAME}" # 恢复备份
        systemctl start "${SERVICE_NAME}"
        rm -f "/tmp/${TARBALL_NAME}"
        rm -rf "$UPDATE_TEMP_DIR"
        exit 1
    fi
    
    # 更新 template 文件夹 (如果存在)
    # 这会覆盖 $INSTALL_PATH/template 文件夹中的内容
    # $INSTALL_PATH/config 和 $INSTALL_PATH/data 文件夹将因未被操作而保持不变
    if [ -d "$UPDATE_TEMP_DIR/template" ]; then
        echo -e "${GREEN_COLOR}更新 template 文件夹到 ${INSTALL_PATH} ...${RES}"
        # 为了安全地覆盖，可以先删除旧的 template 目录 (如果确定这是期望行为)
        # 或者直接用 cp -Rf (但要小心符号链接等)
        # 更安全的方式是，如果 INSTALL_PATH/template 存在，先删除，再复制新的
        if [ -d "$INSTALL_PATH/template" ]; then
            rm -rf "$INSTALL_PATH/template" || echo -e "${YELLOW_COLOR}警告：删除旧的 template 文件夹失败，可能导致更新不完全。${RES}"
        fi
        if ! cp -R "$UPDATE_TEMP_DIR/template" "$INSTALL_PATH/"; then
            echo -e "${YELLOW_COLOR}警告：复制新的 template 文件夹失败。二进制文件已更新，但模板可能未更新。${RES}"
            # 根据需要，这里也可以决定是否回滚整个更新
        fi
    else
        echo -e "${YELLOW_COLOR}新版本压缩包中未找到 template 文件夹，跳过更新 template。${RES}"
    fi

    # 注意：这会覆盖用户在 $INSTALL_PATH/config 中的任何手动修改。
    # 如果您希望保留用户修改，则不应执行此操作，或者需要更复杂的合并逻辑。
    # 以下操作会覆盖现有的 config 目录。
    if [ -d "$UPDATE_TEMP_DIR/config" ]; then
        echo -e "${GREEN_COLOR}更新 config 文件夹到 ${INSTALL_PATH} ...${RES}"
        if [ -d "$INSTALL_PATH/config" ]; then
            rm -rf "$INSTALL_PATH/config" || echo -e "${YELLOW_COLOR}警告：删除旧的 config 文件夹失败，可能导致更新不完全。${RES}"
        fi
        if ! cp -R "$UPDATE_TEMP_DIR/config" "$INSTALL_PATH/"; then
            echo -e "${YELLOW_COLOR}警告：复制新的 config 文件夹失败。二进制文件已更新，但配置可能未更新。${RES}"
        fi
    else
        echo -e "${YELLOW_COLOR}新版本压缩包中未找到 config 文件夹，跳过更新 config。${RES}"
    fi

    echo -e "${GREEN_COLOR}设置新版 ${BINARY_NAME} 执行权限...${RES}"
    chmod +x "$INSTALL_PATH/${BINARY_NAME}"

    # 清理临时文件和备份
    rm -f "/tmp/${TARBALL_NAME}"
    rm -rf "$UPDATE_TEMP_DIR"
    rm -f "/tmp/${BINARY_NAME}.bak"

    echo -e "${GREEN_COLOR}启动 ${APP_NAME} 服务...${RES}"
    if systemctl start "${SERVICE_NAME}"; then
        echo -e "${GREEN_COLOR}${APP_NAME} 更新完成并已启动！${RES}"
    else
        echo -e "${RED_COLOR}${APP_NAME} 更新完成，但服务启动失败。请检查日志: sudo journalctl -u ${SERVICE_NAME}${RES}"
    fi
}


UNINSTALL() {
    if [ ! -f "$INSTALL_PATH/${BINARY_NAME}" ] && [ ! -d "$INSTALL_PATH" ]; then # Check both binary and directory
         handle_error 1 "未在 $INSTALL_PATH 找到 ${APP_NAME} 或相关安装目录，无需卸载。"
    fi
    echo -e "${RED_COLOR}警告：卸载后将删除 ${APP_NAME} 安装目录 (${INSTALL_PATH})、数据（如果存储在其中）、systemd服务及命令行工具！${RES}"
    read -p "是否确认卸载？[Y/n]: " choice

    case "${choice:-Y}" in # Default to Y if user just presses Enter
        [yY])
            echo -e "${GREEN_COLOR}开始卸载 ${APP_NAME} ...${RES}"
            
            echo -e "${GREEN_COLOR}停止并禁用 ${APP_NAME} 服务...${RES}"
            systemctl stop "${SERVICE_NAME}" >/dev/null 2>&1
            systemctl disable "${SERVICE_NAME}" >/dev/null 2>&1
            
            echo -e "${GREEN_COLOR}删除 ${APP_NAME} 文件和目录...${RES}"
            rm -rf "$INSTALL_PATH"
            rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
            systemctl daemon-reload
            
            if [ -f "$MANAGER_PATH" ] || [ -L "$COMMAND_LINK" ]; then
                echo -e "${GREEN_COLOR}删除命令行工具...${RES}"
                rm -f "$MANAGER_PATH" "$COMMAND_LINK" || {
                    echo -e "${YELLOW_COLOR}警告：删除命令行工具失败，请手动删除：${RES}"
                    echo -e "${YELLOW_COLOR}1. $MANAGER_PATH${RES}"
                    echo -e "${YELLOW_COLOR}2. $COMMAND_LINK${RES}"
                }
            fi
            
            echo -e "${GREEN_COLOR}${APP_NAME} 已完全卸载。${RES}"
            ;;
        *)
            echo -e "${GREEN_COLOR}已取消卸载。${RES}"
            ;;
    esac
}

SHOW_MENU() {
    # Update INSTALL_PATH based on current service file, important if user installs to custom path then runs menu
    CURRENT_INSTALL_PATH=$(GET_INSTALLED_PATH)
    # Check if binary actually exists at this path for menu options that need it
    local app_is_installed=false
    if [ -f "${CURRENT_INSTALL_PATH}/${BINARY_NAME}" ]; then
        app_is_installed=true
    fi

    echo -e "\n欢迎使用 ${APP_NAME} 管理脚本 (v1.0.0)\n"
    echo -e "${GREEN_COLOR}当前检测到的安装路径: ${CURRENT_INSTALL_PATH}${RES}"
    if ! $app_is_installed && [ "$CURRENT_INSTALL_PATH" != "/opt/${APP_NAME}" ]; then
         echo -e "${YELLOW_COLOR}警告: ${BINARY_NAME} 未在上述路径找到，某些操作可能不可用或针对默认路径。${RES}"
    elif ! $app_is_installed; then
         echo -e "${YELLOW_COLOR}${APP_NAME} 似乎未安装。${RES}"
    fi
    echo ""
    echo -e "${GREEN_COLOR}1、安装 ${APP_NAME}${RES}             (默认路径: /opt/${APP_NAME} 或指定)"
    if $app_is_installed; then
        echo -e "${GREEN_COLOR}2、更新 ${APP_NAME}${RES}             (从 ${CURRENT_INSTALL_PATH})"
        echo -e "${GREEN_COLOR}3、卸载 ${APP_NAME}${RES}             (从 ${CURRENT_INSTALL_PATH})"
        echo -e "${GREEN_COLOR}-------------------------${RES}"
        echo -e "${GREEN_COLOR}4、查看状态${RES}"
        # Option 5 for Reset Password removed, add if your app supports it
        echo -e "${GREEN_COLOR}-------------------------${RES}"
        echo -e "${GREEN_COLOR}6、启动 ${APP_NAME}${RES}"
        echo -e "${GREEN_COLOR}7、停止 ${APP_NAME}${RES}"
        echo -e "${GREEN_COLOR}8、重启 ${APP_NAME}${RES}"
        echo -e "${GREEN_COLOR}9、查看日志${RES}"
    else
        echo -e "${YELLOW_COLOR}2、更新 ${APP_NAME}${RES}             (请先安装)"
        echo -e "${YELLOW_COLOR}3、卸载 ${APP_NAME}${RES}             (请先安装)"
        echo -e "${GREEN_COLOR}-------------------------${RES}"
        echo -e "${YELLOW_COLOR}4、查看状态${RES}               (请先安装)"
        echo -e "${GREEN_COLOR}-------------------------${RES}"
        echo -e "${YELLOW_COLOR}6、启动 ${APP_NAME}${RES}             (请先安装)"
        echo -e "${YELLOW_COLOR}7、停止 ${APP_NAME}${RES}             (请先安装)"
        echo -e "${YELLOW_COLOR}8、重启 ${APP_NAME}${RES}             (请先安装)"
        echo -e "${YELLOW_COLOR}9、查看日志${RES}               (请先安装)"
    fi
    echo -e "${GREEN_COLOR}-------------------------${RES}"
    echo -e "${GREEN_COLOR}10、安装/更新命令行工具${RES}"
    echo -e "${GREEN_COLOR}0、退出脚本${RES}"
    echo
    read -p "请输入选项 [0-10]: " choice

    # Use CURRENT_INSTALL_PATH for operations if app is installed
    if $app_is_installed; then
        INSTALL_PATH="$CURRENT_INSTALL_PATH"
    fi

    case "$choice" in
        1)
            echo -e "请输入 ${APP_NAME} 的安装路径 (直接回车将使用默认路径: /opt/${APP_NAME}):"
            read -r custom_path
            if [ -n "$custom_path" ]; then
                 # Re-run the INSTALL_PATH setting logic with the custom path
                if ! [[ $custom_path == */${APP_NAME} ]]; then
                    INSTALL_PATH="$custom_path/${APP_NAME}"
                else
                    INSTALL_PATH="$custom_path"
                fi
                parent_dir=$(dirname "$INSTALL_PATH")
                if [ ! -d "$parent_dir" ]; then
                    mkdir -p "$parent_dir" || handle_error 1 "无法创建目录 $parent_dir"
                fi
                if ! [ -w "$parent_dir" ]; then
                    handle_error 1 "目录 $parent_dir 没有写入权限"
                fi
            else
                INSTALL_PATH="/opt/${APP_NAME}" # Reset to default for new install
            fi
            CHECK "install" # Pass context to CHECK
            INSTALL
            INIT
            SUCCESS
            return 0 ;;
        2)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装，无法更新。${RES}"; return 1; fi
            UPDATE
            return 0 ;; # UPDATE exits on its own usually
        3)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装，无法卸载。${RES}"; return 1; fi
            UNINSTALL
            return 0 ;; # UNINSTALL exits on its own
        4)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装。${RES}"; return 1; fi
            if systemctl is-active "${SERVICE_NAME}" >/dev/null 2>&1; then
                echo -e "${GREEN_COLOR}${APP_NAME} 当前状态为：运行中${RES}"
                systemctl status "${SERVICE_NAME}" --no-pager | grep -E "Active:|Loaded:|Main PID:|Tasks:|Memory:|CGroup:"
            else
                echo -e "${RED_COLOR}${APP_NAME} 当前状态为：已停止或未运行${RES}"
                if systemctl list-units --full -all | grep -q "${SERVICE_NAME}.service"; then
                    systemctl status "${SERVICE_NAME}" --no-pager | grep -E "Active:|Loaded:"
                else
                    echo -e "${YELLOW_COLOR}服务文件 ${SERVICE_NAME}.service 可能不存在。${RES}"
                fi
            fi
            return 0 ;;
        # Option 5 (Reset Password) removed
        6)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装。${RES}"; return 1; fi
            echo -e "${GREEN_COLOR}正在启动 ${APP_NAME} ...${RES}"
            if systemctl start "${SERVICE_NAME}"; then echo -e "${GREEN_COLOR}${APP_NAME} 已启动${RES}"; else echo -e "${RED_COLOR}${APP_NAME} 启动失败，请查看日志。${RES}"; fi
            return 0 ;;
        7)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装。${RES}"; return 1; fi
            echo -e "${GREEN_COLOR}正在停止 ${APP_NAME} ...${RES}"
            if systemctl stop "${SERVICE_NAME}"; then echo -e "${GREEN_COLOR}${APP_NAME} 已停止${RES}"; else echo -e "${RED_COLOR}${APP_NAME} 停止失败。${RES}"; fi
            return 0 ;;
        8)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装。${RES}"; return 1; fi
            echo -e "${GREEN_COLOR}正在重启 ${APP_NAME} ...${RES}"
            if systemctl restart "${SERVICE_NAME}"; then echo -e "${GREEN_COLOR}${APP_NAME} 已重启${RES}"; else echo -e "${RED_COLOR}${APP_NAME} 重启失败，请查看日志。${RES}"; fi
            return 0 ;;
        9)
            if ! $app_is_installed; then echo -e "${RED_COLOR}错误: ${APP_NAME} 未安装。${RES}"; return 1; fi
            echo -e "${GREEN_COLOR}显示最近的 ${SERVICE_NAME} 日志 (按 q 退出):${RES}"
            journalctl -u "${SERVICE_NAME}" -e --no-pager -n 50
            return 0 ;;
        10)
            INSTALL_CLI
            return 0 ;;
        0)  exit 0 ;;
        *)  echo -e "${RED_COLOR}无效的选项${RES}"; return 1 ;;
    esac
}

# 主程序逻辑
if [ $# -eq 0 ]; then
    while true; do
        clear
        SHOW_MENU
        # Wait for user input if an error occurred or for a short period on success
        if [ $? -eq 0 ]; then
            echo -e "\n${GREEN_COLOR}操作完成。按任意键返回主菜单...${RES}"
        else
            echo -e "\n${RED_COLOR}操作出现错误。按任意键返回主菜单...${RES}"
        fi
        read -n 1 -s -r # Wait for a single key press, silent, raw
    done
elif [ "$1" = "install" ]; then
    # For command line install, $2 can be the path
    # The INSTALL_PATH logic at the beginning handles $2 if provided
    CHECK "install"
    INSTALL
    INIT
    SUCCESS
    exit 0
elif [ "$1" = "update" ]; then
    if [ $# -gt 1 ]; then
        echo -e "${RED_COLOR}错误：update 命令不需要指定路径${RES}"
        echo -e "正确用法: sudo $0 update"
        exit 1
    fi
    UPDATE
    exit 0
elif [ "$1" = "uninstall" ]; then
    if [ $# -gt 1 ]; then
        echo -e "${RED_COLOR}错误：uninstall 命令不需要指定路径${RES}"
        echo -e "正确用法: sudo $0 uninstall"
        exit 1
    fi
    UNINSTALL
    exit 0
else
    echo -e "${RED_COLOR}错误的命令: $1${RES}"
    echo -e "用法: sudo $0 [command]"
    echo -e "命令:"
    echo -e "  install [安装路径]    # 安装 ${APP_NAME} (默认路径: /opt/${APP_NAME})"
    echo -e "  update              # 更新 ${APP_NAME}"
    echo -e "  uninstall           # 卸载 ${APP_NAME}"
    echo -e "  <没有参数>          # 显示交互菜单"
    exit 1
fi