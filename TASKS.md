Bugs:

1. "View Club Players" always give you an empty page. // DONE
2. "Get Rating History", search for "C0101-10077486" returns a table with a lot of N/A // DONE
3. Performance Bug for "Search Players", it tooks very long and run often into timeouts. Search for exact string instead of prefix / suffix pattern. // DONE
4. Searching for players with 'Tran' you see for Player ID a lot of UNKNOWN-10799083, why UNKNOWN? // DONE
5. Player ID after - is a 3 digit number, not like this C0101-10077486. // DONE
6. Player with ID C0327-10224943 shows as inactive, but actually this player is active. // DONE
7. Show Players, doesn't matter where, always with First Name and Last Name, instead of only First Name. // DONE
8. Searching for Reuther or other longer names runs into 'Search failed: Failed to fetch' // DONE
9. If searching for Club ID 'C0327', the players were shown. If chosing sort by DWZ Rating or Birth Year, I got error "Club players lookup failed: Club or players not found". // DONE
10. Never shows not active players or players not belonging to a Team in /demo. So i.e. UNKNOWN-10224937 should never be shown. For the REST-API itself do not add an extra route, but add active flag and set is as default. // DONE

// FIXED - Test failures resolved:
11. CSV format test failures - Fixed SendCSVResponse validation and data handling // DONE
12. Empty club ID test expectation - Correctly routes to SearchClubs, not GetClub // DONE  
13. E2E test localization - Updated tests to expect German error messages // DONE

Missing Features:

1. In the demo site there is no pagination for all html pages. So whenever you find more than the number of players/clubs/tournament for the first page, you cannot go to the next page. // DONE
2. Translate all pages for /demo to German. So "normal" users would see German only. Keep everything else belonging to a developer like REST-API documentation, swagger, code etc. in English. // DONE 
3. For Club players list, the default sort order of the player is DWZ, since it the internal ranking of the club directly. Club players should be sorted by DWZ rating (strongest first). // DONE
4. Showing player details for a certain tournement like this page https://www.schachbund.de/turnier/C531-634-S25.html is missing. So we know details about the tournement itself for the moment. But we do not know about which players change the DWZ how and also the tournament details with its result per player like this: https://www.schachbund.de/turnier/C531-634-S25/Ergebnisse.html // DONE