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
# 未复用tcp连接的情况，不会被识别到，但是每次请求都需要重新建立tcp连接
#
########################################################################################################################

import requests

url = 'http://127.0.0.1:8080'  # 你请求的URL

# 发送100次请求
for seqNum in range(1, 31):
    response = requests.get(url)
    print(f"{seqNum}: Status Code: {response.status_code}, response = {response.text}")  # 打印状态码
