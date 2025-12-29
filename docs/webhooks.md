# Webhook Setup

## GitHub (GitHub App)
1. Create a GitHub App in your org/user settings.
2. Set the webhook URL to `https://<your-domain>/webhooks/github`.
3. Set a webhook secret and export it as `GITHUB_WEBHOOK_SECRET`.
4. Subscribe to the events you need (e.g., pull requests, push).
5. Deploy Githooks and ensure the endpoint is reachable over HTTPS.

## GitLab
1. In your GitLab project/group, go to **Settings → Webhooks**.
2. Set the URL to `https://<your-domain>/webhooks/gitlab`.
3. Add a secret token and export it as `GITLAB_WEBHOOK_SECRET` (optional).
4. Select the events you want (e.g., Merge request events, Push events).
5. Save and test delivery.

## Bitbucket (Cloud)
1. In your Bitbucket repo, go to **Repository settings → Webhooks**.
2. Set the URL to `https://<your-domain>/webhooks/bitbucket`.
3. Set a UUID secret and export it as `BITBUCKET_WEBHOOK_SECRET` (optional).
4. Select the events you want (e.g., Pull request created, Repo push).
5. Save and test delivery.
