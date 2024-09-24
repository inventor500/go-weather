# Weather

This is a simple program that retrieves the current weather forecast and prints it to the terminal.

Data is sourced from <https://forecast.weather.gov>.

## Useage

### Command Line Parameters
| *Parameter* | *Description*                                                         |
|-------------|-----------------------------------------------------------------------|
| `-zip`      | The zip code the query for. This also accepts "city, state" locations |
| `-u`        | User agent to use                                                     |
| `-c`        | Location of the config file                                           |

If the `-zip` parameter is not specified, then the program will look for a zip code at the end of the parameter list.  
Example: `weather 53226`

### Configuration File
The configuration file is located at `$XDG_CONFIG_HOME/weather/config.json` by default.

| *Key*              | *Description*                 |
|--------------------|-------------------------------|
| `default-location` | Default location to query for |
| `user-agent`       | User agent to use for queries |

### Environment Variables
| *Variable*   | *Description*     |
|--------------|-------------------|
| `USER_AGENT` | User agent to use |

## Notes

- The User Agent string priority is from most specific to least specific. It first checks for a value specific to this run (the command line parameter), then for specific to this program (the config file), then for specific to this user/machine (the environment variable. If none of these are found, it pretends to be Firefox 130 on a Windows 10 computer.
