# Hood
 ## Palmer
 **2** hours before and after civil twilight
 OR 
 when moon is gibbous or full.
`{{ or (betweenRiseSet .Now .Astro 2) (brightMoon .Astro) }}`

 ## Vista
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

# Three Sisters
 ## Bachelor
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

# Rainier
 ## Paradise
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

 ## Schurman
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

 ## Crystal Mountain
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

# Shasta
 ## Snowcrest
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

# Whitney
 ## Lone Pine
 civil twilight (I think their cam is set up like this as well.)
`{{ betweenRiseSet .Now .Astro 0 }}`

# Fuji
 ## North
 1 hour before and after civil twilight
`{{ betweenRiseSet .Now .Astro 1 }}`

# Mont Blanc
 ## tourism
 civil twilight (webcam isn't really that good)
`{{ betweenRiseSet .Now .Astro 0 }}`

# Eiger
 ## Wixi
 civil twilight (cam actually seems to do 30 mins AFTER sunrise/set
`{{ betweenRiseSet .Now .Astro 0 }}`