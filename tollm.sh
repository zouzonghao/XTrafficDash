#!/bin/sh

# --- 配置区 ---

# 输出文件名
OUTPUT_FILE="code_summary_for_llm.txt"

# 要包含的文件扩展名列表（用空格分隔）
# 你可以根据需要添加或删除扩展名
FILE_EXTENSIONS=".go .sql .py .js .java .html .css .ts  .yml .yaml .json .xml .c .cpp .h .hpp .cs .rb .php .rs .kt .vue"

# 要忽略的目录列表（使用 shell 数组，每个元素是一个目录）
# 每个目录都应该用引号括起来，以防万一路径包含空格或特殊字符
EXCLUDE_DIRS=("./web/node_modules" "./web/dist" ) # <-- 修改此处为数组

# --- 脚本逻辑 ---

echo "开始收集代码文件..."
# 使用 printf 来正确格式化数组元素的显示
echo "目标文件扩展名: $(echo "$FILE_EXTENSIONS" | tr ' ' '\n')"
echo "要忽略的目录:"
printf "  %s\n" "${EXCLUDE_DIRS[@]}"
echo "输出文件: ${OUTPUT_FILE}"
echo ""

# 1. 清空或创建输出文件
# 使用 > "$OUTPUT_FILE" 来覆盖现有文件，如果文件不存在则创建它
> "$OUTPUT_FILE"
if [ $? -ne 0 ]; then
    echo "错误：无法写入输出文件 '$OUTPUT_FILE'。请检查权限。"
    exit 1
fi

# 2. 构建 find 命令的参数

# 构建排除目录的 find 参数
EXCLUDE_ARGS=""
for dir in "${EXCLUDE_DIRS[@]}"; do # <-- 使用数组遍历
    # 确保目录非空且存在，避免不必要的警告
    if [ -n "$dir" ]; then
        # -path "$dir" -prune: 匹配到该目录时，不再深入查找（prune）
        # -o: OR 逻辑，继续处理其他文件
        # 注意：这里用 printf"%q" 来安全的引用目录名，以防万一。
        EXCLUDE_ARGS="$EXCLUDE_ARGS -path $(printf "%q" "$dir") -prune -o"
    fi
done

# 构建查找文件的 find 参数
FIND_ARGS=""
FIRST_EXT=true
for ext in $FILE_EXTENSIONS; do
    # 确保扩展名非空
    if [ -n "$ext" ]; then
        if [ "$FIRST_EXT" = true ]; then
            # 使用 printf"%q" 来安全的引用扩展名模式
            FIND_ARGS="-name $(printf "%q" "*$ext")"
            FIRST_EXT=false
        else
            FIND_ARGS="$FIND_ARGS -o -name $(printf "%q" "*$ext")"
        fi
    fi
done

# 检查是否找到了任何扩展名
if [ -z "$FIND_ARGS" ]; then
    echo "错误：未配置任何要包含的文件扩展名。"
    exit 1
fi

# 组合查找和排除的 find 命令
# .                  : 从当前目录开始查找
# \( $EXCLUDE_ARGS -false \) : 这是一个组合条件，用于处理排除的目录。
#                           -path "$dir" -prune -o 会为每个排除目录生成 `-path "dir" -prune -o`。
#                           最后的 `-false` 结合 `-o`，使得排除目录的整个表达式为真（但由于 prune，不会深入）。
# -o                 : OR 逻辑，将排除条件和查找条件分开
# -type f            : 只查找普通文件
# \( $FIND_ARGS \)  : 查找符合设定扩展名的文件
# -print0            : 使用 null 字符作为分隔符打印文件名，安全处理包含特殊字符的文件名
FIND_COMMAND="find . \( $EXCLUDE_ARGS -false \) -o -type f \( $FIND_ARGS \) -print0"

# echo "调试：find 命令: $FIND_COMMAND" # 如果需要调试，取消此行注释

# 3. 遍历找到的文件并处理
# 使用 eval 来正确执行包含特殊字符（如引号）的 FIND_COMMAND
# 使用 while IFS= read -r -d $'\0' 来安全地逐行读取 find 的输出
eval "$FIND_COMMAND" | while IFS= read -r -d $'\0' file; do
    echo "正在处理: $file"

    # 添加文件路径和分隔符到输出文件
    echo "--- 文件路径 (File Path): $file ---" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE" # 在路径后添加空行

    # 添加文件内容到输出文件
    # 使用 cat 读取文件内容并追加 >>
    if cat "$file" >> "$OUTPUT_FILE"; then
        # 成功读取并追加内容
        echo "" >> "$OUTPUT_FILE" # 在内容后添加空行
        echo "--- 文件内容结束 (End of Content) ---" >> "$OUTPUT_FILE"
    else
        # 读取文件时可能发生错误 (例如，权限问题)
        echo "" >> "$OUTPUT_FILE" # 依然添加空行
        echo "[错误: 无法读取文件内容 - $file]" >> "$OUTPUT_FILE"
        echo "--- 文件内容结束 (End of Content) ---" >> "$OUTPUT_FILE"
    fi
     # 在每个文件条目之间添加一个空行以增加可读性
    echo "" >> "$OUTPUT_FILE"

done

# 检查 find 命令本身是否出错（例如，无法访问某些目录）
# 注意：这只能捕获 find 命令本身的错误，无法捕获 while 循环内部的 cat 错误（已在循环内处理）
# && [ ! -s "$OUTPUT_FILE" ] 是用来判断如果 find 命令本身出错，但输出文件是空的，则显示更明确的警告。
if [ $? -ne 0 ] && [ ! -s "$OUTPUT_FILE" ]; then
     echo "警告：`find` 命令执行时可能遇到问题，或者没有找到匹配的文件。"
fi

echo "----------------------------------------"
echo "代码文件汇总完成！"
echo "结果已保存到: ${OUTPUT_FILE}"
echo "请检查 '$OUTPUT_FILE' 文件。"
echo "提示：如果文件非常大，请注意可能产生的文本文件大小。"

exit 0
