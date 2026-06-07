# MahJong LTS
福州麻将的重制版，未来会长期更新支持

# 即刻体验官方版本
[https://mahjong.wepayto.win](https://mahjong.wepayto.win)

# 技术栈
- MikuTap
- Gin
- Vue
- SSE

# TODO
- ✅ S3 兼容存储以存储可以选择的角色
- 显示 IP 归属地
- 云端数据库以启用 LoginRequire
- KV 缓存数据库查询结果

## 麻将规则
- ✅福州麻将

# 自己部署
## S3 兼容存储
需要准备角色素材，S3 存储桶内应有以下四个文件夹
- head
- full
- confirm
- critical

`head` 角色头像，`full` 角色全身像格式应为 `webp`，`confirm` 为选择角色时触发的语音，`critical`为触发吃碰杠后播放的特效，格式应为支持透明(Alpha)通道的 `webm`。绑定一个域名并设置 CORS 规则。

注意使用 Safai WebView 的 iOS 系列浏览器对 webm 的支持并不友好，所以禁用了特效。

### 示例
`Persona` 语境下的 `Joker` 的头像。
```
head/Persona/Joker.webp
```
![https://cm.mahjong.wepayto.win/head/Persona/Joker.webp](https://cm.mahjong.wepayto.win/head/Persona/Joker.webp)

`Persona` 语境下的 `Joker` 的暴击特效，注意这个特效是有声音的。
```
critical/Persona/Joker.webm
```
[https://cm.mahjong.wepayto.win/critical/Persona/Joker.webm](https://cm.mahjong.wepayto.win/critical/Persona/Joker.webm)

## 环境变量
### S3 兼容存储
创建一个权限为`管理员只读`的 API 令牌，配置环境变量
- `S3_ACCESS_KEY_ID` 
- `S3_SECRET_ACCESS_KEY`
- `S3_ENDPOINT`     用以执行 ListObject 操作，一般为云服务商提供
- `S3_BUCKET`
- `S3_ACCESS_POINT` 用以直接访问其中内容，一般为自己添加
- `S3_REGION` (可选)












