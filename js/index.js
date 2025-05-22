searchingPaintingsMode = true;
showLoginMode = true;

//html-elements
var loginForm;
var signupForm;
var authorizedWindow;

var authorized = false;

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


async function search()
{
    if(searchingPaintingsMode)
    {
        fetch_paintings();
    }            
    else
    {
        fetch_authors();
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
    const responce = await fetch(`http://localhost:8080/paintings_${query}`)
    if (!responce.ok)
    {
        //check for errors later
        node.innerHTML = ''
        return;
    }
    const data = await responce.json()
    console.log(data.length)


    innerHypertext = ''
    for (let i = 0; i < data.length; i++)
    {
        where_to_find = data[i].where_to_find==""?"не экспонируется":data[i].where_to_find
        innerHypertext += `
        <div class="search-res">
            <div class="search-pic">
                <img src="assets/${data[i].picture_address}" width="100%">
            </div>
            <div class="search-desc"> 
                НАЗВАНИЕ: <b>${data[i].title}(${data[i].creation_year})</b><br>
                АВТОР: <b>${data[i].author_name}</b><br>
                МУЗЕЙ: <b><a>${data[i].museum_name}</b></a><br>
                ГДЕ НАЙТИ: <b>${where_to_find}</b>
                <hr>
                <button onclick="likePainting(${data[i].id})">мне нравится</button>
                <button onclick="dislikePainting(${data[i].id})">мне не нравится</button>
                <hr>
            </div>
        </div>
        `
    };
    console.log(innerHypertext)
    console.log(data.length)
    node.innerHTML = innerHypertext;
}

async function fetch_authors() {
    var node = document.getElementById('paintings_placeholder');
    const query = document.getElementById("searchbar").value
    if(query == "")
    {
        node.innerHTML = ''
        return
    }
    const responce = await fetch(`http://localhost:8080/authors_${query}`)
    if (!responce.ok)
    {
        //check for errors later
        node.innerHTML = ''
        return;
    }
    const data = await responce.json()
    console.log(data.length)
    
    innerHypertext = ''
    for (let i = 0; i < data.length; i++)
    {
        innerHypertext += `<div>${data[i].name}(${data[i].birth_year}-${data[i].death_year})<hr></div>`
    };
    node.innerHTML = innerHypertext;
}

async function likePainting(paintingId)
{
    console.log(paintingId)
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
        console.log("GOOD");
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
        console.log("GOOD");
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
        document.getElementById("change-mode").innerHTML = `<input type="button" value="search authors instead" onclick="changeMode()">`;
        document.getElementById("searchbar").placeholder = "Find your favorite paintings...";
    }
    else
    {
        document.getElementById("change-mode").innerHTML = `<input type="button" value="search paintings instead" onclick="changeMode()">`;
        document.getElementById("searchbar").placeholder = "Find your favorite authors...";
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
        if(data.length == 0)
        {
            favoritesList.innerHTML = `Кажется, у вас пока нет любимых картин...`
        }
        else
        {
            innerHypertext = ``
            for(let i = 0; i < data.length; i++)
            {
                innerHypertext += 
                `
                    <div>
                        <div class="search-pic">
                            <img src="assets/${data[i].picture_address}" width="100%">
                        </div>
                        <div class="search-desc"> 
                            <b>${data[i].title}(${data[i].creation_year})</b><br>
                            АВТОР: <b>${data[i].author_name}</b><br>
                            МУЗЕЙ: <b><a>${data[i].museum_name}</b></a><br>                            
                        </div>
                    </div>
                    <hr>
                `;
            }
            favoritesList.innerHTML = innerHypertext;
        }
    }
    else
    {
        favoritesList.innerHTML = `Кажется, у вас пока нет любимых картин...`
    }  
}
