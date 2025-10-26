# VHS tape for gittui demo - theme toggling
Output gittui-demo.gif

# Reasonable resolution with small font to show full app
Set Width 1920
Set Height 1080
Set FontSize 14
Set Theme "Dracula"
Set TypingSpeed 100ms
Set Shell "bash"

# Build if needed
Hide
Type "go build -o gittui ."
Enter
Sleep 2
Type "clear"
Enter
Show

# Launch gittui
Type "./gittui"
Enter
Sleep 5

# Hold down 't' key for 10 seconds - rapid fire theme cycling
Type@1ms "tttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt"

# Pause to show final theme
Sleep 3

# Quit
Type "q"
Sleep 0.5

# Cleanup
Hide
Type "clear"
Enter
