#!/bin/bash

DB="backend/xtrafficdash.db"

# 随机生成IP
rand_ip() {
  echo "$((RANDOM%223+1)).$((RANDOM%255)).$((RANDOM%255)).$((RANDOM%255))"
}

# 随机生成端口
rand_port() {
  echo "$((RANDOM%55535+10000))"
}

# 随机邮箱
rand_email() {
  echo "user$((RANDOM%1000000))@example.com"
}

# 生成30天日期（2025-06-23 ~ 2025-07-22）
DATES=(
  "2025-06-23" "2025-06-24" "2025-06-25" "2025-06-26" "2025-06-27"
  "2025-06-28" "2025-06-29" "2025-06-30" "2025-07-01" "2025-07-02"
  "2025-07-03" "2025-07-04" "2025-07-05" "2025-07-06" "2025-07-07"
  "2025-07-08" "2025-07-09" "2025-07-10" "2025-07-11" "2025-07-12"
  "2025-07-13" "2025-07-14" "2025-07-15" "2025-07-16" "2025-07-17"
  "2025-07-18" "2025-07-19" "2025-07-20" "2025-07-21" "2025-07-22"
)

# 2GB~20GB 高质量随机下载（openssl+bc）
gen_random_down() {
  min=2147483648
  max=21474836480
  range=$((max - min + 1))
  # openssl 生成8字节二进制，转为十进制
  n=$(openssl rand -hex 8)
  # 只取前15位，防止溢出
  n=${n:0:15}
  # 用 bash 10进制处理（去掉前导0和非数字）
  n=$((10#${n//[^0-9]/}))
  echo $(( min + n % range ))
}
# 9%~11% 随机上传
gen_random_up() {
  local down=$1
  percent=$(( 9 + RANDOM % 3 ))
  echo $(( down * percent / 100 ))
}

for ((svc=1; svc<=2; svc++)); do
  IP=$(rand_ip)
  NAME="测试节点$RANDOM"
  sqlite3 $DB "INSERT INTO services (ip_address, custom_name, first_seen, last_seen, status) VALUES ('$IP', '$NAME', '2025-06-23 00:00:00', '2025-07-22 23:59:59', 'active');"
  SERVICE_ID=$(sqlite3 $DB "SELECT id FROM services WHERE ip_address='$IP' ORDER BY id DESC LIMIT 1;")

  # 随机2个端口
  declare -a PORTS
  declare -a INBOUND_IDS
  for ((p=1; p<=2; p++)); do
    PORT=$(rand_port)
    TAG="inbound-$PORT"
    sqlite3 $DB "INSERT INTO inbound_traffics (service_id, tag, port, last_updated, status) VALUES ($SERVICE_ID, '$TAG', $PORT, '2025-07-22 23:59:59', 'active');"
    INBOUND_ID=$(sqlite3 $DB "SELECT id FROM inbound_traffics WHERE service_id=$SERVICE_ID AND tag='$TAG' ORDER BY id DESC LIMIT 1;")
    PORTS[$p]=$PORT
    INBOUND_IDS[$p]=$INBOUND_ID
  done

  # 随机2个用户
  declare -a USERS
  declare -a CLIENT_IDS
  for ((u=1; u<=2; u++)); do
    EMAIL=$(rand_email)
    sqlite3 $DB "INSERT INTO client_traffics (service_id, email, last_updated, status) VALUES ($SERVICE_ID, '$EMAIL', '2025-07-22 23:59:59', 'active');"
    CLIENT_ID=$(sqlite3 $DB "SELECT id FROM client_traffics WHERE service_id=$SERVICE_ID AND email='$EMAIL' ORDER BY id DESC LIMIT 1;")
    USERS[$u]=$EMAIL
    CLIENT_IDS[$u]=$CLIENT_ID
  done

  # 填充端口流量历史
  for ((p=1; p<=2; p++)); do
    INBOUND_ID=${INBOUND_IDS[$p]}
    PORT=${PORTS[$p]}
    TAG="inbound-$PORT"
    for ((i=0;i<30;i++)); do
      DAY=${DATES[$i]}
      DOWN=$(gen_random_down)
      UP=$(gen_random_up $DOWN)
      sqlite3 $DB "INSERT INTO inbound_traffic_history (inbound_traffic_id, service_id, tag, date, daily_up, daily_down, created_at) VALUES ($INBOUND_ID, $SERVICE_ID, '$TAG', '$DAY', $UP, $DOWN, '2025-07-22 23:59:59');"
    done
  done

  # 填充用户流量历史
  for ((u=1; u<=2; u++)); do
    CLIENT_ID=${CLIENT_IDS[$u]}
    EMAIL=${USERS[$u]}
    for ((i=0;i<30;i++)); do
      DAY=${DATES[$i]}
      DOWN=$(gen_random_down)
      UP=$(gen_random_up $DOWN)
      sqlite3 $DB "INSERT INTO client_traffic_history (client_traffic_id, service_id, email, date, daily_up, daily_down, created_at) VALUES ($CLIENT_ID, $SERVICE_ID, '$EMAIL', '$DAY', $UP, $DOWN, '2025-07-22 23:59:59');"
    done
  done

done

echo "测试数据已写入 $DB"