# screenpos

A simple way to get a position on your screen using your keyboard and the visual aid of a grid.

For use with tools such as `xdotool` allowing you to click on the selected position.

## Use with xdotool

`xdotool mousemove $(screenpos) click 1`

## Directions for use

To use `screenpos` effectively, it is recommended that you install `xdotool` and use the command above. Include this in a shell script and bind the execution of the script to a key combination. How to do this will vary depending on your desktop environment.

Decide on the position you want to click and press the key combination to activate `screenpos`. The right-angled arrows in the top left of cells indicate the position that will be clicked, you may find that one of these is already in the position you want to click. If this is the case, simply type the two characters appearing in the cell. If not, you may translate the grid slightly using your arrow keys to move the closest arrow into the correct position.

## Configuration

`screenpos` is configurable through a JSON config file. This is optional.

Use `-c myconfig.json` or `--config myconfig.json` flags to use your custom config.

Example config:

```json
{
    "arrowColour": "#0080ff",
    "arrowSize": 5,
    "arrowOpacity": 255,
    "fontColour": "#ffffff",
    "fontDropColour": "#000000",
    "fontOpacity": 255,
    "lineColour": "#ffffff",
    "lineDropColour": "#000000",
    "lineOpacity": 64,
    "gridStep": 6
}
```

## Demos

### Navigating file explorer

![](demo1.gif)

### Navigating a website

![](demo2.gif)

### Navigating code

![](demo3.gif)
