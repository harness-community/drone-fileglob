# drone-findfiles

A plugin to search for files using relative or absolute paths and filter based on [Ant path pattern](https://ant.apache.org/manual/dirtasks.html#patterns).

To learn how to utilize Drone plugins in Harness CI, please consult the provided [documentation](https://developer.harness.io/docs/continuous-integration/use-ci/use-drone-plugins/run-a-drone-plugin-in-ci).

## Usage

The following settings changes this plugin's behavior.

* ```filter```: Ant style pattern to search for files. For example, ```**/*.txt``` searches for all ```.txt``` files in directories.
* ```excludes``` (optional): Pattern to exclude files from the search result. For example, ```**/*.zip``` excludes files with zip extension from the result.
* ```dir``` (optional) : Directory in which to perform the search, if not specificed use the current directory.

## Output

The search result is output as a JSON with the following properties.

* ```name```: The file name.
* ```path```: The complete path to the file.
* ```isDirectory```: A boolean to indicate if the path refer to a directory or not.
* ```length```: The length in bytes of the file.
* ```lastModified```: The last modified formatted as RFC3339.

Below is an example of the output when run the plugin using this code repository directory.

```json
[
    {
        "name": "main.go",
        "path": "drone-findfiles/main.go",
        "isDirectory": false,
        "length": 1130,
        "lastModified": "2024-09-12T19:45:00Z"
    },
    {
        "name": "pipeline.go",
        "path": "drone-findfiles/plugin/pipeline.go",
        "isDirectory": false,
        "length": 5424,
        "lastModified": "2024-09-12T19:45:00Z"
    },
    {
        "name": "plugin.go",
        "path": "drone-findfiles/plugin/plugin.go",
        "isDirectory": false,
        "length": 3444,
        "lastModified": "2024-09-12T19:45:00Z"
    },
    {
        "name": "plugin_test.go",
        "path": "drone-findfiles/plugin/plugin_test.go",
        "isDirectory": false,
        "length": 9838,
        "lastModified": "2024-09-12T19:45:00Z"
    }
]
```

The plugin uses the Drone environment variable ```DRONE_OUTPUT``` to write the search result.

## Step Definition

Below is an example to use the plugin inside a Harness CI pipeline.

```yaml
- step:
    type: Plugin
    name: Find Files
    identifier: Find_Files
    spec:
      connectorRef: harness-docker-connector
      image: harness-community/drone-findfiles:linux-amd64
      settings:
        filter: "**/*.go"
```

To use the output JSON inside the Harness pipeline use the expression ```<+steps.STEP_ID.output.outputVariables.FILES_INFO>``` where the ```STEP_ID``` is the identifier from the find files plugin. 

You can use the expression and inspect each attribute using ```jq``` tool inside a Run step.

```bash
files_info='<+steps.Find_Files.output.outputVariables.FILES_INFO>'

echo "Path"
echo $files_info | jq '.[].path'
```

## Building

Build the plugin binary:

```text
scripts/build.sh
```

## Testing

Execute the plugin from your current working directory:

```bash
PLUGIN_FILTER="**/*.yml" \
PLUGIN_EXCLUDES="**/b.*" \
DRONE_OUTPUT="drone_output.properties" \
go run main.go
```
