const controllers = [];

function formatIcon(name, isDir) {
    if (isDir) return "📁";

    const ext = name.split('.').pop().toLowerCase();

    const iconMap = {
        // 文档
        'txt': '📄',
        'md': '📝',
        'pdf': '📕',
        'doc': '📄',
        'docx': '📄',
        'xls': '📊',
        'xlsx': '📊',
        'ppt': '📽️',
        'pptx': '📽️',

        // 图片
        'jpg': '🖼️',
        'jpeg': '🖼️',
        'png': '🖼️',
        'gif': '🖼️',
        'bmp': '🖼️',
        'svg': '🖼️',
        'webp': '🖼️',

        // 视频
        'mp4': '🎬',
        'mkv': '🎞️',
        'avi': '📼',
        'mov': '🎞️',

        // 音频
        'mp3': '🎵',
        'wav': '🎶',
        'flac': '🎶',

        // 代码
        'js': '📜',
        'ts': '📜',
        'go': '🐹',
        'py': '🐍',
        'java': '☕',
        'c': '💻',
        'cpp': '💻',
        'cs': '💻',
        'html': '🌐',
        'css': '🎨',
        'json': '🧾',
        'sh': '🖥️',
        'bat': '🖥️',

        // 压缩包
        'zip': '🗜️',
        'rar': '🗜️',
        '7z': '🗜️',
        'gz': '🗜️',
        'tar': '🗜️',
        // 默认
        'default': '📄'
    };

    return iconMap[ext] || iconMap['default'];
}

function render(path = "") {
    const list = document.getElementById("file-list");
    list.dataset.path = path;

    fetch("/list?path=" + encodeURIComponent(path))
        .then(res => res.json())
        .then(data => {
            list.innerHTML = "";
            if(!data){return}
            data.forEach(item => {
                const div = document.createElement("div");
                div.className = "file-item";

                const left = document.createElement("div");
                left.className = "file-left";

                const icon = document.createElement("span");
                icon.textContent = formatIcon(item.name,item.is_dir);
                icon.className = "file-icon";

                const name = document.createElement("span");
                name.textContent = item.name;
                name.className = item.is_dir ? "dir" : "";

                if (item.is_dir) {
                    name.onclick = () => render(item.path);
                }

                left.appendChild(icon);
                left.appendChild(name);

                const right = document.createElement("div");
                console.log(item)
                //item.is_dir ||
                if (!item.size) {
                    const button = document.createElement("button");
                    button.textContent = "计算大小";
                    button.className = "button";
                    button.onclick = (e) => {
                        e.stopPropagation();
                        button.disabled = true;
                        button.textContent = "计算中...";

                        const controller = new AbortController();
                        controllers.push(controller);

                        fetch("/size?path=" + encodeURIComponent(item.path), { signal: controller.signal })
                            .then(res => res.text())
                            .then(size => {
                                button.textContent = size;
                                button.disabled = false;
                                const index = controllers.indexOf(controller);
                                if (index > -1) controllers.splice(index, 1);
                            })
                            .catch(err => {
                                if (err.name === 'AbortError') {
                                    button.textContent = "已取消";
                                } else {
                                    button.textContent = "失败";
                                }
                                button.disabled = false;
                                const index = controllers.indexOf(controller);
                                if (index > -1) controllers.splice(index, 1);
                            });
                    };
                    right.appendChild(button);
                } else {
                    right.textContent = item.size;
                }

                div.appendChild(left);
                div.appendChild(right);
                list.appendChild(div);
            });
        });
}

window.onload = () => {
    render("/");
    // 返回上级目录
    document.getElementById("back-btn").onclick = () => {
        const path = document.getElementById("file-list").dataset.path || "/";
        const parts = path.split("/").filter(Boolean);
        if (parts.length === 0) return; // 根目录不返回
        parts.pop();
        const parentPath = "/" + parts.join("/");
        render(parentPath);
    };

    // 批量计算所有“计算大小”按钮
    document.getElementById("batch-btn").onclick = () => {
        document.querySelectorAll(".file-item button").forEach(btn => {
            if (btn.textContent === "计算大小") {
                btn.click();
            }
        });
    };
};

// 页面卸载时取消所有未完成请求
window.addEventListener("beforeunload", () => {
    controllers.forEach(controller => controller.abort());
});
