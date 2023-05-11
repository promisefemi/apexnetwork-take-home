Apex Network Take Home.

NB: Since game sessions could persist after dice roll completion, each dice roll is a session and the first dice roll holds the winning value, not the game itself. 

Apis:

| Path                | Method | Description                     | 
|---------------------| ---- |---------------------------------|
| /register           | POST | Register new user               |
| /fund-wallet        | POST | Fund user wallet                |
| /get-wallet-balance | GET | Get user wallet and details     |
| /roll-dice          | POST | Roll dice in a game | 
| /end-game           | POST | End all games and dice rolls |
| /start-game | POST | Start a new game |
| /check-active-game | GET | Check if there is an active game in progress |
| /transactions | GET | Get all user transactions |

Repo contains Postman collection for test.

File Structure:

/handler/handler.go -- Contains all api endpoints

/util/util.go - Contains helpers and utility functions

/main.go - Entry point for application