# chzzk-go

[Chzzk](https://chzzk.naver.com) 공식 API를 위한 Go 클라이언트 라이브러리입니다.

[English](../README.md)

## 설치

`go get github.com/sdkim96/chzzk-go`

## 사용법

```go

import chzzk "github.com/sdkim96/chzzk-go"

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
    defer cancel()

    c := chzzk.New(nil).WithAPIKey("your-chzzk-api-key-here")
    user, err := c.User.Me(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Println("ChannelID: ", user.ChannelID)
    fmt.Println("ChannelName: ", user.ChannelName)
}
```

이 예제를 사용하려면 치지직에 애플리케이션을 등록해야 합니다.

### 애플리케이션 등록

브라우저에서 https://developers.chzzk.naver.com 에 접속합니다.

![alt text](image.png)

아래와 같이 애플리케이션을 등록합니다.

![alt text](image1.png)

> 주의: 리다이렉션 URL은 반드시 `http://localhost:57777/callback`으로 입력해야 합니다.

포트가 `57777`로 고정된 이유: https://github.com/sdkim96/chzzk-go/blob/main/internal/login/login.go#L9-L17

등록이 완료되면 아래와 같이 확인할 수 있습니다.

![alt text](image2.png)

### 로그인

배포된 바이너리를 사용하면 간편하게 로그인할 수 있습니다.

```bash
go install github.com/sdkim96/chzzk-go/cmd/chzzk-login@latest
chzzk-login
```

또는 [GitHub Releases](https://github.com/sdkim96/chzzk-go/releases)에서 빌드된 바이너리를 다운로드할 수 있습니다.

CLI가 OAuth 플로우를 안내하고, 완료되면 액세스 토큰을 출력합니다.

## 인증

두 가지 인증 방식을 지원합니다.

### Client Credentials

서버 간 API 호출(세션 인증, 토큰 관리 등)에 사용합니다.

```go
c := chzzk.New(nil).WithClientAuth("your-client-id", "your-client-secret")
```

### API Key (사용자 액세스 토큰)

사용자 범위 API 호출(사용자 정보 조회, 채팅 구독 등)에 사용합니다.

```go
c := chzzk.New(nil).WithAPIKey("your-access-token")
```

## 서비스

### Token

```go
c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)

// 새 토큰 발급
token, err := c.Token.New(ctx, chzzk.TokenNewRequest{
    TokenRequest: chzzk.TokenRequest{
        GrantType:    chzzk.GrantTypeAuthorizationCode,
        ClientID:     clientID,
        ClientSecret: clientSecret,
    },
    Code:  code,
    State: state,
})

// 토큰 갱신
token, err := c.Token.Refresh(ctx, chzzk.TokenRefreshRequest{
    TokenRequest: chzzk.TokenRequest{
        GrantType:    chzzk.GrantTypeRefreshToken,
        ClientID:     clientID,
        ClientSecret: clientSecret,
    },
    RefreshToken: "your-refresh-token",
})

// 토큰 폐기
err := c.Token.Revoke(ctx, chzzk.RevokeTokenRequest{
    ClientID:      clientID,
    ClientSecret:  clientSecret,
    Token:         "token-to-revoke",
    TokenTypeHint: "access_token",
})
```

### User

```go
c := chzzk.New(nil).WithAPIKey("your-access-token")

user, err := c.User.Me(ctx)
fmt.Println(user.ChannelID, user.ChannelName)
```

### Channel

```go
// 채널 정보 조회 (WithClientAuth)
c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)
channels, err := c.Channel.Get(ctx, "channelId1", "channelId2")

// 매니저, 팔로워, 구독자 (WithAPIKey)
c := chzzk.New(nil).WithAPIKey("your-access-token")

managers, err := c.Channel.Managers(ctx)
followers, nextPage, err := c.Channel.Followers(ctx, nil, nil)
subscribers, nextPage, err := c.Channel.Subscribers(ctx, nil, nil, nil)
```

### Live

```go
// 라이브 목록 조회 (WithClientAuth)
c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)
lives, next, err := c.Live.Get(ctx, nil, nil)

// 스트림키, 방송 설정 (WithAPIKey)
c := chzzk.New(nil).WithAPIKey("your-access-token")
key, err := c.Live.Key(ctx)
setting, err := c.Live.Setting(ctx)

// 방송 설정 변경
title := "내 방송"
err := c.Live.PatchSetting(ctx, &chzzk.PatchLiveSettingRequest{
    Title:    &title,
    Category: &chzzk.Category{Type: "GAME", ID: "League_of_Legends"},
})
```

### Category

```go
c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)
categories, err := c.Category.Search(ctx, "리그", nil)
```

### Session (실시간 이벤트)

```go
c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)

// 세션 URL 획득
sessionURL, err := c.Session.AuthClient(ctx)

// 채팅 이벤트 구독/구독 해제
err := c.Session.SubscribeChat(ctx, sessionKey)
err := c.Session.UnSubscribeChat(ctx, sessionKey)
```

### Socket.IO

`socketio` 패키지는 실시간 이벤트 수신을 위한 Socket.IO v2 클라이언트를 제공합니다.

```go
import "github.com/sdkim96/chzzk-go/socketio"

conn := socketio.New(wsURL,
    socketio.WithHandler("CHAT", func(data []byte) error {
        fmt.Println("Chat:", string(data))
        return nil
    }),
    socketio.WithHandler("DONATION", func(data []byte) error {
        fmt.Println("Donation:", string(data))
        return nil
    }),
)

err := conn.Dial(ctx)
defer conn.Close(ctx, 1000, "done")

err = conn.Loop(ctx) // 컨텍스트가 취소되거나 에러가 발생할 때까지 블로킹
```

## 테스트

```bash
# 유닛 테스트
go test ./...

# 통합 테스트 (CHZZK_CLIENT_ID, CHZZK_CLIENT_SECRET 환경변수 필요)
go test -tags=integration ./...
```

## 라이선스

MIT
