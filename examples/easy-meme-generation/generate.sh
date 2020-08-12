memes="memes.txt"

IFS=";"
while read -a line
do
    impacc -i in/${line[0]} -o out/${line[1]} -t "${line[2]}" -b "${line[3]}"
done < "$memes"
