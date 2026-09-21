# Quickstart

1. Create a WHOOP developer application.
2. Register `http://localhost:8400/callback` as its redirect URI.
3. Configure `WHOOP_CLIENT_ID` and `WHOOP_CLIENT_SECRET`, or store them in the secure backend described in [Authentication](auth.md).
4. Check configuration:

   ```sh
   whoop config diagnose --json
   ```

5. Authorize once:

   ```sh
   whoop auth
   ```

6. Read data:

   ```sh
   whoop recovery --json
   whoop workouts --plain
   whoop weekly
   ```

The CLI requests only `read:profile read:body_measurement read:cycles read:recovery read:sleep read:workout offline`. It does not control devices or write account data. See [Automation](automation.md) before embedding it in another program.
