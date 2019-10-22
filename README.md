# HGNotify

age
@HGNotify [options] [GroupName] [mentions...]

##Summary
The purpose of this bot is to @/mention groups of people in Google's Hangouts Chat by using user created groups, since gchat doesn't seem to already have this functionality. A more broad description of the bot is a group manager.

##LIMITATION **Please read**
**Due to a chat bot in Google Chat not being able to add users to a room, if you use a group in a room where the group members are not already added, they will not be mentioned, you'll just see their user id. Its gravely disapointing, but the result of a limitation within Google's Hangouts Chat system.**

##Examples
**Mentions:** "HEY! @HGNotify HG6, great job on that new product!" would turn into "HEY! @Alexander Wilcots @Robert Rabel @Robert Stone @James Frotten @Cai Black @Taylor Mitchell @Srimathy Thyagarajan, great job on that new product"

**Creating a group:** "@HGNotify create HG1 @Brandon Husbands"

**Making a group private:** "@HGNotify restrict HG1"

**Adding a group member:** "@HGNotify add HG1 @Taylor Mitchell"

**Removing a group member:** "@HGNotify remove HG6 @Robert Stone"

**Delete a group:** "@HGNotify disband Umbrella"

##Options
**create groupName [mentions]**
Create a group containing mentioned members. While I'm not sure why you would, you can initialize an empty group.

**add groupName mentions**
Add mentioned members to the specified GroupName. This can only be used for groups that already exist. If you intend to create a new group use create.

**remove groupName mentions**
Remove mentioned members from the specified GroupName.

**disband groupName**
Delete a group. CAUTION: This can be done to a group containing members. I'd recommend only using delete when necessary.

**restrict groupName**
Toggles group privacy, this disallows any interaction with the group outside the room it was restricted in. (Default: Public)

**list [groupName]**
If used with no groupName, you will receive a list of all groups you can currently use. This will not show any private group that you do not have access to. If used with a groupName, you will see more information about the group specified.

**groupName**
  Replaces groupName with mentions for the group members along with the surrounding message.

**help|usage**
Reprint's this message

##Notes
- Group Names are case insensative.
- Group Names can contain letters, numbers, underscores, and dashes maximum length is 40 characters
- When managing groups, "@HGNotify" must be the first thing in the messages
- When notifying a group the text "@HGNotify GroupName" will be replaced with the members of the group. Just a heads up, so be sure to place that where you'd like it to appear.

- Any problems, comments, or suggestions please send me a message in gchat or email me at alexander.wilcots@endurance.com
