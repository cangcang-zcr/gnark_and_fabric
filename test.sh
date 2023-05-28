#!/bin/bash

# 并发命令数
concurrent_commands=50

# 要执行的命令
command_to_run="curl -s -X GET \"http://127.0.0.1:9000/storepubkey?addressA=张三&addressB=李四&PubKey=xxx\" -H \"Content-Type: application/json\""

# 记录开始时间
start_time=$(date +%s)

# 使用函数并在后台执行命令
execute_command() {
  eval "$command_to_run"
}

# 用于跟踪后台进程的数组
pids=()

# 启动并发命令
for i in $(seq 1 $concurrent_commands); do
  execute_command &
  pids+=($!)
done

# 等待所有后台进程完成
for pid in ${pids[@]}; do
  wait $pid
done

# 记录结束时间
end_time=$(date +%s)

# 计算并显示总时间
total_time=$((end_time - start_time))
echo "总时间: $total_time 秒"
