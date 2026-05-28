"""
自动生成 Docsify 侧边栏 _sidebar.md

用法：
  python gen_sidebar.py

功能：
  - 扫描 docs/ 下的子目录，读取每个 .md 文件的一级标题作为链接文字
  - 自动生成根目录和 docs/ 目录的 _sidebar.md
  - 删除子目录级别多余的 _sidebar.md（Docsify 会自动回退到上级）
  - 新增 .md 文件后重新运行此脚本即可更新侧边栏，无需手动编辑

配置说明：
  - GROUP_ORDER：侧边栏分组的显示顺序
  - GROUP_NAMES：目录名 → 分组显示名的映射（新增目录时在此添加）
"""

import os
import re

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
DOCS_DIR = os.path.join(BASE_DIR, 'docs')

# ── 配置 ──

# 分组在侧边栏中的显示顺序（未列出的目录按字母序排在末尾）
GROUP_ORDER = ['目录1', '目录2']

# 目录名 → 侧边栏分组显示名
GROUP_NAMES = {
    '目录1': '目录1',
    '目录2': '目录2',
}

# 不出现在侧边栏中的文件
SKIP_FILES = {'_sidebar.md', '_navbar.md'}


def get_first_heading(filepath):
    """读取 .md 文件的第一行 # 标题，用作侧边栏链接文字"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            for line in f:
                match = re.match(r'^#\s+(.+)', line)
                if match:
                    return match.group(1).strip()
                # 遇到非空非标题行就停止（标题一定在最开头）
                if line.strip() and not line.startswith('#'):
                    break
    except Exception as e:
        print(f'  警告：无法读取 {filepath}: {e}')
    return None


def generate_sidebar_content():
    """扫描 docs/ 子目录，生成 _sidebar.md 的文本内容"""
    lines = []

    # 收集 docs/ 下的子目录
    subdirs = []
    for name in os.listdir(DOCS_DIR):
        path = os.path.join(DOCS_DIR, name)
        if os.path.isdir(path):
            subdirs.append(name)

    # 按配置顺序排列，未配置的按字母序排在末尾
    ordered = [d for d in GROUP_ORDER if d in subdirs]
    remaining = sorted([d for d in subdirs if d not in GROUP_ORDER])
    all_dirs = ordered + remaining

    for dirname in all_dirs:
        dirpath = os.path.join(DOCS_DIR, dirname)
        display_name = GROUP_NAMES.get(dirname, dirname)

        # 找到目录中所有 .md 文件（排除跳过列表中的文件）
        md_files = []
        for filename in sorted(os.listdir(dirpath)):
            if not filename.endswith('.md'):
                continue
            if filename in SKIP_FILES:
                continue

            filepath = os.path.join(dirpath, filename)
            heading = get_first_heading(filepath)
            link_name = filename[:-3]  # 去掉 .md 后缀

            if heading:
                link_text = heading
            else:
                # 没有标题时用文件名作为显示文字
                link_text = link_name

            link_url = f'/docs/{dirname}/{link_name}'
            md_files.append((link_text, link_url))

        if not md_files:
            print(f'  跳过空目录: {dirname}')
            continue

        # 写入分组（用空行分隔 → loose list → 折叠插件正常工作）
        lines.append(f'- {display_name}')
        for link_text, link_url in md_files:
            lines.append(f'  - [{link_text}]({link_url})')
        lines.append('')  # 组间空行，保证 loose list

    return '\n'.join(lines)


def remove_per_dir_sidebars():
    """删除子目录级别的 _sidebar.md，让 Docsify 回退到上级"""
    for name in os.listdir(DOCS_DIR):
        dirpath = os.path.join(DOCS_DIR, name)
        if not os.path.isdir(dirpath):
            continue
        sidebar_file = os.path.join(dirpath, '_sidebar.md')
        if os.path.exists(sidebar_file):
            os.remove(sidebar_file)
            print(f'  删除: {sidebar_file}')


def main():
    print('正在扫描 docs/ 目录...')

    content = generate_sidebar_content()

    if not content.strip():
        print('错误：没有找到任何 .md 文件，请检查 docs/ 目录')
        return

    # 写入根目录 _sidebar.md
    root_sidebar = os.path.join(BASE_DIR, '_sidebar.md')
    with open(root_sidebar, 'w', encoding='utf-8') as f:
        f.write(content + '\n')
    print(f'  生成: {root_sidebar}')

    # 写入 docs/ 目录 _sidebar.md（首页路由优先加载此文件）
    docs_sidebar = os.path.join(DOCS_DIR, '_sidebar.md')
    with open(docs_sidebar, 'w', encoding='utf-8') as f:
        f.write(content + '\n')
    print(f'  生成: {docs_sidebar}')

    # 删除子目录级别的 _sidebar.md
    print('正在清理子目录级别的 _sidebar.md...')
    remove_per_dir_sidebars()

    print('')
    print('侧边栏自动生成完成！')
    print('新增 .md 文件后重新运行 gen_sidebar.py 即可更新。')
    print('新增目录时，在脚本顶部的 GROUP_ORDER 和 GROUP_NAMES 中添加配置。')


if __name__ == '__main__':
    main()
