# service tests

## replay_test.go (pure timer model, table-driven via TestReplay)
- **fresh session splits L evenly**: 600s over 5 actions gives a 120s countdown.
- **a done resets the countdown using the new totals**: after one done at 60s, 4 actions share the remaining 540s (135s each).
- **the countdown keeps running down after a done**: later in the same countdown, the left value shrinks while the total stays.
- **working fast earns more time per action**: finishing early leaves a longer per-action countdown than on-pace work.
- **missing a deadline counts one miss...**: the action stays outstanding and the next countdown is shorter.
- **several misses while the app was closed**: replaying a long gap counts every deadline that fell inside it.
- **a miss before a done**: a deadline passed before the next event is counted, and the done still resets the countdown.
- **the 1-second floor keeps misses finite near L** and **after the floor-limited misses...**: near the end the countdown stops shrinking, so misses stop and time_up follows.
- **reaching L with work outstanding is time_up**: a deadline landing on L is not an extra miss.
- **logging the last action is completed**: completion wins mid-countdown and records the event time.
- **undo restores a count** / **undo after undo** / **a single undo**: undo reverses the newest not-yet-undone done and resets the countdown.
- **pause freezes both clocks**, **resume starts a fresh countdown**, **paused time does not count**, **misses are counted in active time after a pause**: pausing stops all progress.
- **end event abandons**: an end event ends the session at its own time.
- **same-timestamp events order by ID**: ties are broken by event ID, regardless of input order.
- **events after the session has ended are ignored**.

## session_test.go (stateful service over an in-memory store with a fake clock)
- **TestGetDraft_CreatesOnceSeededFromLastSessionMinutes**: only one draft exists and it is seeded from the saved length.
- **TestAddChore_ValidatesAndTrimsAndAppends**: blank names and counts below one are rejected; names are trimmed; chores append in order.
- **TestDraftChoreEditing_UpdateRemoveReorder**: draft chores can be edited, removed and reordered, and a partial reorder is rejected.
- **TestChoresAreEditableOnlyOnADraft**: editing or removing a chore of a started session is rejected.
- **TestSetDraftLength_StoresMinutesAsSeconds**: the length is stored in seconds; negatives are rejected.
- **TestStartSession_RequiresChoresAndPositiveLength**: starting needs a chore and a length above zero.
- **TestStartSession_SavesLastSessionMinutesAndReturnsFreshSnapshot**: starting saves the length for next time and returns the even-split snapshot.
- **TestStartSession_RejectedWhileAnotherSessionIsRunning**: only one session runs at a time.
- **TestGetRunningSession_NilWhenNothingRuns**: no running session gives nil, not an error.
- **TestLogAction_ReducesOutstandingAndResetsCountdown**: logging an action updates the totals and resets the countdown.
- **TestLogAction_RejectsChoreWithNothingOutstandingOrUnknown**: finished and unknown chores are rejected.
- **TestLogAction_RequiresARunningSession**: logging with nothing running is rejected.
- **TestCompletingLastActionPersistsEndOnce**: completion writes end status and the completion time, not the later read time.
- **TestTimeUpWhileClosedIsDetectedPersistedAndReportedOnce**: a session that ran out unseen ends as time_up, is reported once, then is no longer running.
- **TestMissesWhileAppWasClosedAreCountedByANewService**: a new service over the same store counts missed deadlines.
- **TestLogActionAfterTimeRanOutIsRejected**: actions after time-up are rejected.
- **TestUndo_RestoresLastActionAndRejectsWhenNothingToUndo**: undo works once per logged action and is rejected otherwise.
- **TestPauseResume_RejectDoubleAndFreezeTime**: double pause and resume-while-running are rejected, and paused time is not charged.
- **TestEndSession_AbandonsAndPersists**: ending early stores abandoned and a second end is rejected.
- **TestGetSummary_RejectsDraftsAndUnknownSessions**: summaries need a started, existing session.
- **TestRepeatSession_CopiesListAndReplacesExistingDraftChores**: repeating copies names, counts, order and length into the one draft.
- **TestRecentChoreNames_CappedAtLimitNewestFirst**: quick picks are capped and newest first.
- **TestUpdateSettings_ValidatesAndPersists**: settings round-trip and a zero length is rejected.
- **TestFullSession_FileBackedSurvivesReopen**: a session in a file-backed DB keeps its log and counts misses after reopening.
