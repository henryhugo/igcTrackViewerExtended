## igcTrackViewerExtended

Work realized by Hugo HENRY
Assignement 2
Develop an online service that will allow users to browse information about IGC files.

# Features not realized

MongoDB storage, GET /api/ticker/<timestamp>, problem with the clock trigger


# URLs

- https://igctrackviewerextended.herokuapp.com/paragliding use GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api use GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api/track/ use POST with the write json format  or GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api/track/id0 use GET and format id12, id20

- https://igctrackviewerextended.herokuapp.com/paragliding/api/track/id0/pilot use GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api/ticker/latest use GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api/ticker use GET

- https://igctrackviewerextended.herokuapp.com/paragliding/api/webhook/new_track/ use POST with the write json format 

- https://igctrackviewerextended.herokuapp.com/paragliding/api/webhook/new_track/id0 use GET to see the webook or DELETE to delete the webhook

- https://igctrackviewerextended.herokuapp.com//admin/api/tracks_count use GET to see the number of tracks

- https://igctrackviewerextended.herokuapp.com//admin/api/tracks use DELETE to delete all tracks


