# Instruction of chzzk-login

## Prerequisite

1. You must have created your Chzzk application. If not, you can find the cookbook [here](https://github.com/sdkim96/chzzk-go#how-to-make-application).

2. You must have set up the following credentials before running this program.
- CHZZK_CLIENT_ID (client id)
- CHZZK_CLIENT_SECRET (client secret)

3. You can either:
- register the preceding values to environment variables as they are.
- memorize them.

## Steps

### Installation

You can install the prebuilt binary from [Github Releases](https://github.com/sdkim96/chzzk-go/releases).

If you have Go installed, you can easily obtain it by executing this command:

```bash
go install github.com/sdkim96/chzzk-go/cmd/chzzk-login@latest
chzzk-login
```

## Instruction

1. Execute the binary.

```bash
sungdongkim@Sungdongui-Macmini chzzk-go_0.1.0_darwin_amd64 % ls
LICENSE		README.md	chzzk-login
sungdongkim@Sungdongui-Macmini chzzk-go_0.1.0_darwin_amd64 % ./chzzk-login
```

2. Type 'y' when prompted to continue.

```bash
=== chzzk-go login ===

Before you start, make sure you have:
  1. Created an app at https://developers.chzzk.naver.com
  2. Registered the redirect URL: http://localhost:57777/callback

Continue? (y/n): y
```

3. Open the URL in the browser.

```bash
Client ID: ************ (from env)
Client Secret: ********** (from env)

--- Step 1: Authorize ---
Open this URL in your browser:

  https://chzzk.naver.com/account-interlock?clientId=******a&redirectUri=http://localhost:57777/callback&state=chzzk-go-login-1783339781

Waiting for callback...
```

4. If successful, the callback server will respond as "Login successful. You may close this window."

5. You can copy the output from console.

```bash
Authorization code received.

--- Step 2: Exchange Token ---
Token received.

=== Login Complete ===

  Access Token:  
  ********
  Refresh Token: 
  ****************
  Expires In:    86400s

Example usage:

  c := chzzk.New(nil).WithAPIKey("**********")
sungdongkim@Sungdongui-Macmini chzzk-go_0.1.0_darwin_amd64 % 
```

6. That's it.