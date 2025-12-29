# Webhook Setup

Use provider-native webhook configuration to point to the Githooks endpoints.

## GitHub (GitHub App)
1. Create a GitHub App in your org/user settings.
2. Set webhook URL to `https://<your-domain>/webhooks/github`.
3. Set `GITHUB_WEBHOOK_SECRET` in your environment.
4. Subscribe to the events you need.
5. Deploy behind HTTPS.

## GitLab
1. Go to **Settings → Webhooks** in your project/group.
2. Set URL to `https://<your-domain>/webhooks/gitlab`.
3. Set `GITLAB_WEBHOOK_SECRET` (optional).
4. Select the events you want.
5. Save and test delivery.

## Bitbucket (Cloud)
1. Go to **Repository settings → Webhooks**.
2. Set URL to `https://<your-domain>/webhooks/bitbucket`.
3. Set `BITBUCKET_WEBHOOK_SECRET` (optional, X-Hook-UUID).
4. Select the events you want.
5. Save and test delivery.
