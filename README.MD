# myproxy
参考 [lightsocks](https://github.com/gwuhaolin/lightsocks)
### 基于socks5实现的翻墙工具
#### 1.使用方式
#####  1.配置文件

```json
{
  "remote_addr": "127.0.0.1", 
  "remote_port": 8888,
  "local_addr": "0.0.0.0",
  "local_port": 1080,
  "password": "fiRJM4paHRRU9rvlBuyfnuq/xI/++FyoaCVQvP2dR3eSRNzyGrrQEq1ua869H7GCZ8cIEJuieK5jnOCagXusG2oJwISJlA8APW2MekvL5MaOQfM1iC5ig7URZibhvvFZAdg+8PtF7gTa/B6pT+3IB8J0fTKkC9LKtzk4zTaRycFeX7inUe+YDXXXhfo3YClMdpegWwI6GErjxaOQgIddKG8rFwq0lQzp/9aN9Og/mU7UYffnQqHMFqaGz9WTLyxVVlOW3ldGDnNpubCrMQUw9Tx5syeyUkh8KsMZ5hVDZK+qIPkcIXBNNN0tE1hyZd/ibIvbQNlx638jA7bRpdMiOw==",
  "localhost": "127.0.0.1"
}
```
#### 2.TODO

- [x] 开发完成可以使用
- [ ] PAC支持
- [x] HTTPS 支持
