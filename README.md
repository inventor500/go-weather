# Weather

This is a simple program that retrieves the current weather forecast and prints it to the terminal.

Data is sourced from <https://forecast.weather.gov>.

## Useage

**Command line parameters**:
| *Parameter* | *Description*              |
|-------------|----------------------------|
| `-zip`      | The zip code the query for |
| `-u`        | User agent to use          |

If the `-zip` parameter is not specified, then the program will look for a zip code at the end of the parameter list.  
Example: `weather 53226`

**Environment variables**:
| *Variable*   | *Description*                          |
|--------------|----------------------------------------|
| `USER_AGENT` | User agent to use. Overridden by `-u`. |


