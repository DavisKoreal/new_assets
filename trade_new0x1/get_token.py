import json
import requests

def get_token():
    url = "https://api.kucoin.com/api/v1/bullet-public"
    response = json.loads(requests.post(url).text)
    token = response['data']['token']
    #print(token)
    wsurl = response['data']['instanceServers']
    return token

token = get_token()