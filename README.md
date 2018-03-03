# FractalNet

[![Build Status](https://travis-ci.org/fractalbach/fractalnet.svg?branch=master)](https://travis-ci.org/fractalbach/fractalnet)

To Play the current version of whatever happens to be live, go to the
server:

## Location of Server:  http://35.230.55.6/

Hosted on Google Cloud.



# Cellular Automata

* Each grid square is one of 4 different types:  Red, Green, Blue, or Empty.
* The "Neighborhood" of a cell is the 8 surrounding cells.


Color | Enemy | Fun Description
------|-------|------------------------
Red | Blue | Water puts out a fire.
Blue | Green | Earth grows out of water.
Green | Red | Fire burns the earth. 





# Message Examples

Location of WebSocket:
```
ws://35.230.55.6/ws
```


Notes: 

* The *name* of the field **does** matter. 
* The *order* of the fields **does not** matter.  


## Chat

```JSON 
{
    "EventType": "Chat",
    "EventBody": "Hello World!",
}    
```



## Drop Bomb 💣 onto a Square (for the Game of War)

Note: The "Value" field must be either 1 or 2.
It corresponds to Player 1 or Player 2.
Without entering that field, the message might be ignored.


```JSON 
{
    "EventType": "LaBomba",
    "Value": 1,
    "Location": 
    {
        "X": 0,
        "Y": 47,
    }
}    
```



## Update to the Next Generation (Life Update)

```JSON 
{
    "EventType": "LifeUpdate",
}
```




## Directly Change a Square (Life Change)


Currently, the Grid is 48 x 48.  So the Location coordinates _x_ and _y_ are in the range of [0, 47].


Values match with different kinds of squares (can depend on the game). 
For example, in "The Game of War":

* 0: neutral (empty)
* 1: player 1
* 2: player 2
* 3: fire
* 4: fading fire / fallout zone 
* 5: ''' 
* 6: '''


```JSON 
{
    "EventType": "LifeChange",
    "Value": 1,
    "Location": 
    {
        "X": 0,
        "Y": 47,
    }
}    
```


## Directly Change Multiple Squares

When you want many changes to be made within the same game tick:

```JSON 
{
    "EventType": "ChangeMany",
    "Changes": 
    [
        {
            "Value": 1,
            "Location": 
            {
                "X": 1,
                "Y": 10
            }
        },
        {
            "Value": 2,
            "Location": 
            {
                "X": 2,
                "Y": 20
            }
        }
    ]
}   
```








-----------------------------------------------------------------------------







