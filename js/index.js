//=====================SUPPORT-FUNCTIONS-AKA-C-STYLE-INCLUDE-OF-UTILS======================================
function formHighlightedSubstring(initial, toFind)
{
    console.log(initial)
    console.log(toFind)
    index = initial.toLowerCase().indexOf(toFind.toLowerCase())
    return `${initial.substring(0,index)}<b style="background-color:yellow">${initial.substring(index, index+toFind.length)}</b>${initial.substring(index+toFind.length, initial.length)}`
}
//===============================END-OF-INCLUDE-SECTION====================================================

searchingPaintingsMode = true;
showLoginMode = true;

//html-elements
var loginForm;
var signupForm;
var authorizedWindow;

var authorized = false;

paintingsRegistry = new Map();


window.onload=init;
async function init()
{
    
    loginForm = document.getElementById('login-form');
    loginForm.addEventListener('submit', login);
    signupForm = document.getElementById('signup-form');
    signupForm.addEventListener('submit', signup);
    authWindow = document.getElementById("auth");

    var authorized = false;
    authorizedStr = localStorage.getItem("authorized")
    if(authorizedStr == "true")
    {
        authorized = true;
    }

    if(authorized == true)
    {
        console.log("authorized now")
        const response = await fetch(`/login_info`, {credentials: 'include'});
        if (response.ok)
        {
            const data = await response.json()
            console.log(data)
            changeToAuthorized(data);
        }
        else
        {
            authorized = false;
            localStorage.setItem("authorized", authorized)
        }
    }
}

async function fetch_paintings() {
    var node = document.getElementById('paintings-placeholder');
    const query = document.getElementById("searchbar").value
    if(query == "")
    {
        node.innerHTML = ''
        return
    }
    isNowSearchingByTitle = true
    queryText = ``
    if(searchingPaintingsMode)
    {
        isNowSearchingByTitle = true
        queryText = `/paintings_${query}`
        if(authorized) {
            queryText = `/login_paintings_${query}`
        }
    }
    else
    {
        isNowSearchingByTitle = false
        queryText = `/paintings_by_${query}`
        if(authorized) {
            queryText = `login_paintings_by_${query}`
        }
    }
    
    const response = await fetch(queryText)
    if (!response.ok)
    {
        if(authorized)
        {
            if (response.status == 401)
            {
                logout();
            }
        }
        node.innerHTML = ''
        return;
    }
    const data = await response.json()
    console.log(data.length)


    innerHypertext = ''
    paintingsRegistry.clear();
    for (let i = 0; i < data.length; i++)
    {
        paintingsRegistry.set(data[i].id, data[i]);
        where_to_find = data[i].where_to_find==""?"не экспонируется":data[i].where_to_find
        innerHypertext += `
        <div class="search-res">
            <div class="search-pic">
                <img src="assets/${data[i].picture_address}" width="100%">
            </div>
            <div class="search-desc">`
		if (isNowSearchingByTitle == true)
		{
             innerHypertext += `НАЗВАНИЕ: <b>${formHighlightedSubstring(data[i].title, query)}(${data[i].creation_year})</b><br> 
								АВТОР: <b>${data[i].author_name}</b><br>`
		}
		else
		{
		     innerHypertext += `НАЗВАНИЕ: <b>${data[i].title}(${data[i].creation_year})</b><br> 
								АВТОР: <b>${formHighlightedSubstring(data[i].author_name, query)}</b><br>`
		}
		innerHypertext += `
            МУЗЕЙ: <b><a>${data[i].museum_name}</b></a><br>
        	ГДЕ НАЙТИ: <b>${where_to_find}</b>`;
        if(authorized)
        {
            if(data[i].liked == 1)
            {
                innerHypertext += `
                <hr>
                <div id="like-state${data[i].id}" style="display:none;">${parseInt(data[i].liked)}</div>
                <button id="like-button${data[i].id}" onclick="hitLikeButton(${data[i].id})" style="color:white; background-color:#f567be;">&hearts;</button> мне нравится<br>
                <hr>`
            }
            else
            {
                innerHypertext += `
                <hr>
                <div id="like-state${data[i].id}" style="display:none;">${parseInt(data[i].liked)}</div>
                <button id="like-button${data[i].id}" onclick="hitLikeButton(${data[i].id})" style="color: black; background-color:white;">&hearts;</button> мне нравится<br>
                <hr>`
            }
        }
        innerHypertext += `</div></div>`
    };
    node.innerHTML = innerHypertext;
    console.log(paintingsRegistry)
}

async function hitLikeButton(buttonId)
{
    //console.log(document.getElementById(`like-state${buttonId}`).innerHTML == 1)
    if (parseInt(document.getElementById(`like-state${buttonId}`).innerHTML) == 1)
    {
        dislikePainting(buttonId)
    }
    else
    {
        likePainting(buttonId)
    }
}

async function likePainting(paintingId)
{
    data = {
        painting_id: paintingId,
    };
    console.log(JSON.stringify(data));
    const response = await fetch(`http://localhost:8080/favorite`, {
    method: "POST",
    credentials: 'include',
    body: JSON.stringify(data),
    headers: {
        "Content-type": "application/json; charset=UTF-8"
    }
    });
    if (!response.ok)
    {
        console.log("BAD");
    }
    else
    {
        document.getElementById(`like-state${paintingId}`).innerHTML = 1
        document.getElementById(`like-button${paintingId}`).style.cssText = "color:white; background-color:#f567be;"

        likedObjData = paintingsRegistry.get(paintingId)

		newPaintingTagText = `
        <div id = "favorite-${likedObjData.id}">
            <div class="search-pic">
                <img src="assets/${likedObjData.picture_address}" width="100%">
            </div>
            <div class="search-desc"> 
                <b>${likedObjData.title}(${likedObjData.creation_year})</b><br>
                АВТОР: <b>${likedObjData.author_name}</b><br>
                МУЗЕЙ: <b><a>${likedObjData.museum_name}</b></a><br>                            
            </div>
            <hr>
        </div>`
        favorites = document.getElementById(`cabinet-favorites`)
		boolNoFav = document.getElementById("no-favorites")
		console.log(boolNoFav)
		if(boolNoFav == null)
		{
			favorites.innerHTML += newPaintingTagText;
		}
		else
		{
			favorites.innerHTML = newPaintingTagText;
		}
    }
}

async function dislikePainting(paintingId)
{
    data = {
        painting_id: paintingId,
    };
    console.log(JSON.stringify(data));
    const response = await fetch(`http://localhost:8080/favorite`, {
    method: "DELETE",
    credentials: 'include',
    body: JSON.stringify(data),
    headers: {
        "Content-type": "application/json; charset=UTF-8"
    }
    });
    if (!response.ok)
    {
        console.log("BAD");
    }
    else
    {
        document.getElementById(`like-state${paintingId}`).innerHTML = 0
        document.getElementById(`like-button${paintingId}`).style.cssText = "color:black; background-color:white;"
        document.getElementById(`favorite-${paintingId}`).outerHTML = ``;
		if(document.getElementById(`cabinet-favorites`).innerHTML.replace(/ /g, "").length < 10)
		{
			document.getElementById(`cabinet-favorites`).innerHTML = `<div id = "no-favorites" style="display:none;"></div>Кажется, у вас пока нет любимых картин...`
		}
    }
}



function logout(event)
{
    event.preventDefault();

    loginForm.style.display = 'block';
    signupForm.style.display = 'none';
    authWindow.style.display = 'none';
    document.getElementById('change-registration-mode').innerHTML = `вход<br><input type="button" value="регистрация" onclick="changeRegistrationMode()">`;

    document.getElementById('cabinet-login-name').innerHTML = ``;
    document.getElementById('cabinet-museum-list').innerHTML = ``;
    document.getElementById('cabinet-favorites').innerHTML = ``;

    authorized = false;
    localStorage.setItem("authorized", authorized)
}


async function login(event) 
{
    event.preventDefault();
    // Construct post data
    ent_login = document.getElementById('enter-login');
    ent_password = document.getElementById('enter-password');
    data = {
        login : ent_login.value,
        password : ent_password.value,
    };
    const response = await fetch(`http://localhost:8080/login`, {
    method: "POST",
    credentials: 'include',
    body: JSON.stringify(data),
    headers: {
        "Content-type": "application/json; charset=UTF-8"
    }
    });
    const responseData = await response.json();
    if(response.status == 200)
    {
        const response1 = await fetch(`/login_info`, {credentials: 'include'});
        if (response1.ok)
        {
            const data = await response1.json()
            console.log(data)
            changeToAuthorized(data);
        }
        else
        {
            authorized = false;
            localStorage.setItem("authorized", authorized)
        }  
    }
    else if (response.status == 400)
    {
        document.getElementById("enter-status").innerHTML = responseData.error;
    }
    else
    {
        document.getElementById("enter-status").innerHTML = "что-то пошло не так";
    }
}

async function signup(event) 
{
    event.preventDefault();
    // Construct post data
    ent_login = document.getElementById('signup-login');
    ent_password = document.getElementById('signup-password');
    data = {
        login : ent_login.value,
        password : ent_password.value,
    };
    const response = await fetch(`http://localhost:8080/register`, {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
        "Content-type": "application/json; charset=UTF-8"
    }
    });
    const responseData = await response.json();
    if(response.status == 200)
    {
        console.log("user is registered");
        document.getElementById("signup-status").innerHTML = `пользователь \"${data.login}\" зарегистрирован.`;
        
    }
    else if (response.status == 400)
    {
        document.getElementById("signup-status").innerHTML = responseData.error;
    }
    else
    {
        document.getElementById("signup-status").innerHTML = "что-то пошло не так";
    }
}



function changeMode()
{
    searchingPaintingsMode = !searchingPaintingsMode;
    if(searchingPaintingsMode)
    {
        document.getElementById("change-mode").innerHTML = `<input type="button" value="искать по имени автора" onclick="changeMode()">`;
        document.getElementById("searchbar").placeholder = "искать картины по названию";
    }
    else
    {
        document.getElementById("change-mode").innerHTML = `<input type="button" value="искать по названию" onclick="changeMode()">`;
        document.getElementById("searchbar").placeholder = "искать картины по имени автора";
    }
}


function changeRegistrationMode()
{
    var switcher = document.getElementById('change-registration-mode')
    showLoginMode = !showLoginMode;
    if(showLoginMode)
    {
        loginForm.style.display = 'block';
        signupForm.style.display = 'none';
        switcher.innerHTML = `вход<br><input type="button" value="регистрация" onclick="changeRegistrationMode()">`;
    }
    else
    {
        loginForm.style.display = 'none';
        signupForm.style.display = 'block';
        switcher.innerHTML = `<input type="button" value="вход" onclick="changeRegistrationMode()"><br>регистрация`;
    }
}

async function changeToAuthorized(data)
{
    authorized = true;
    localStorage.setItem("authorized", authorized)
    console.log("Hello");
    loginForm.style.display = 'none';
    signupForm.style.display = 'none';
    authWindow.style.display = 'block';
    var switcher = document.getElementById('change-registration-mode');
    switcher.innerHTML = '';

    //customize cabinet
    myLogin = document.getElementById('cabinet-login-name');
    museumList = document.getElementById('cabinet-museum-list');

    myLogin.innerHTML = data.login;

    innerHypertext = ``;
    for (let i = 0; i < data.museums.length; i++)
    {
        innerHypertext += `<li><a href="/museum${data.museums[i].ID}">${data.museums[i].name}</a></li>`
    }
    if(data.museums.length == 0)
    {
        innerHypertext = `<i>Вы пока не являетеся сотрудником ни одного музея.</i>`
    }
    museumList.innerHTML = innerHypertext;

    favoritesList = document.getElementById("cabinet-favorites")
    const response = await fetch(`/favorite`, {credentials: 'include'});
    if (response.ok)
    {
        const data = await response.json()
        console.log(data)
        if(data == null || data.length == 0)
        {
            favoritesList.innerHTML = `<div id = "no-favorites" style="display:none;"></div>Кажется, у вас пока нет любимых картин...`
        }
        else
        {
            innerHypertext = ``
            for(let i = 0; i < data.length; i++)
            {
                innerHypertext += `
                    <div id = "favorite-${data[i].id}">
                        <div class="search-pic">
                            <img src="assets/${data[i].picture_address}" width="100%">
                        </div>
                        <div class="search-desc"> 
                            <b>${data[i].title}(${data[i].creation_year})</b><br>
                            АВТОР: <b>${data[i].author_name}</b><br>
                            МУЗЕЙ: <b><a>${data[i].museum_name}</b></a><br>                            
                        </div>
                        <hr>
                    </div>`
            }
            favoritesList.innerHTML = innerHypertext;
        }
    }
    else
    {
        favoritesList.innerHTML = `Кажется, у вас пока нет любимых картин...`
    }  
}
