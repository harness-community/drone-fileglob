# Local Execution

A directory to assist on debug local execution/validation. Below you have an useful ```launch.json``` definition for local execution. You can inspect the plugin output in file ```drone_output.properties``` from that same directory.

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "main.go",
            "args": [],
            "env": {
                "PLUGIN_GLOB": "**/*.yml",
                "PLUGIN_EXCLUDES": "**/b.*",
                "PLUGIN_DIR": "${cwd}/local",
                "DRONE_OUTPUT": "${cwd}/drone_output.properties"
            }
        }
    ]
}
```
