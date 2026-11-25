# migrations tests

- **TestRun_IsIdempotentAndSeedsSettings**: Running the schema twice (in memory and on a file) succeeds and keeps a setting the user changed between runs, since Run happens on every app start.
- **TestRun_RejectsChoreCountBelowOne**: Inserting a chore with target_count 0 fails, so the database backs up the service's validation.
