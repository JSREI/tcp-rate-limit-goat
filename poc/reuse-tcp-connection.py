# -*- coding: utf-8 -*-
#
# Copyright (C) 2023 by JSREI
# Full license can be found in the LICENSE file.
#
# Author: CC11001100
# URL: https://github.com/JSREI/tcp-rate-limit-goat
#
# Version: 1.0

########################################################################################################################
#
# 复用了tcp连接的情况，使用了同一个Session的会复用同一个tcp连接，因此就可能被识别到
#
########################################################################################################################

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

# 设置重试策略
retry_strategy = Retry(
    total=3,  # 最多重试3次
    status_forcelist=[429, 500, 502, 503, 504],  # 指定哪些状态码需要重试
)

adapter = HTTPAdapter(max_retries=retry_strategy)
session = requests.Session()

# 为session挂载adapter，指定http和https都使用这个adapter
session.mount("http://", adapter)
session.mount("https://", adapter)

url = 'http://127.0.0.1:8080'  # 你请求的URL

# 发送30次请求
for seqNum in range(1, 31):
    response = session.get(url)
    print(f"{seqNum}: Status Code: {response.status_code}, response = {response.text}")  # 打印状态码

# 关闭Session
session.close()