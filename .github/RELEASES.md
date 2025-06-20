# Automated Monthly Releases

This repository is configured to automatically create monthly releases using GitHub Actions.

## Release Schedule

- **Frequency**: Monthly on the 1st day of each month at 09:00 UTC
- **Tag Format**: `v0.YYYY.MM` (e.g., `v0.2024.12`, `v0.2025.01`)
- **Automation**: Fully automated via GitHub Actions

## How It Works

The monthly release workflow (`.github/workflows/release.yml`) performs the following steps:

1. **Generate Version Tag**: Creates a tag in the format `v0.Year.Month`
2. **Check for Existing Tag**: Prevents duplicate releases for the same month
3. **Create Git Tag**: Tags the current commit with the generated version
4. **Generate Release Notes**: Automatically creates changelog from commits since the last release
5. **Create GitHub Release**: Publishes the release with generated notes

## Manual Triggering

You can manually trigger a release by:

1. Going to the "Actions" tab in GitHub
2. Selecting "Monthly Release" workflow
3. Clicking "Run workflow"

This is useful for testing or creating an immediate release.

## Release Notes

Release notes are automatically generated and include:
- A summary of the release
- List of commits since the last monthly release
- Commit hashes for reference

## Permissions

The workflow requires `contents: write` permission to:
- Create tags
- Push to the repository
- Create releases

This permission is automatically granted via the workflow configuration.

## Customization

To modify the release process:

- **Change Schedule**: Edit the `cron` expression in the workflow file
- **Modify Tag Format**: Update the version generation logic
- **Customize Release Notes**: Modify the release notes generation script
- **Add Assets**: Include build artifacts or additional files in releases

## Troubleshooting

If a monthly release fails:

1. Check the Actions tab for error details
2. Ensure the repository has proper permissions
3. Verify no conflicting tags exist
4. Run the workflow manually to test

## Version History

All releases follow semantic versioning principles with the format `v0.Year.Month`, making it easy to track monthly iterations of the project. 