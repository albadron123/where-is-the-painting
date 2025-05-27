function formHighlightedSubstring(initial, toFind)
{
    index = initial.toLowerCase().indexOf(toFind.toLowerCase())
    return `${initial.substring(0,index)}<b style="background-color:yellow">${initial.substring(index, index+toFind.length)}</b>${initial.substring(index+toFind.length, initial.length)}`
}