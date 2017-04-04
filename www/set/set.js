var url;
if (location.protocol === 'https:') {
    url = 'wss://';
} else {
    url = 'ws://';
}
url += location.host + location.pathname + (location.pathname.endsWith('/') ? 'api/ws' : '/api/ws');
var ws = new WebSocket(url);
console.log(new Date().getTime(), 'connecting');

ws.onopen = function (e) {
    console.log(new Date().getTime(), 'connected');
    if (location.hash && location.hash.length > 1) {
        join(location.hash.split('#')[1]);
    } else {
        join('');
    }
};

function join(id) {
    ws.send(JSON.stringify({Type: "join", Join: id}));
}

var cards = document.getElementById('cards');
var gameId = document.getElementById('gameId');
document.getElementById('nosets').onclick = function () {
    ws.send(JSON.stringify({Type: 'nosets', Version: version}))
};

var nut = document.getElementById('nut');
var diamond = document.getElementById('diamond');
var pill = document.getElementById('pill');

var map = {};
var selected = [];
var startTouch = new Date();
var justDragged = false;

function addEventListeners(position, node) {
    node.addEventListener('touchend', selectHandler(position, node, true));
    node.addEventListener('touchstart', function () {
        startTouch = new Date();
    });
    node.addEventListener('touchmove', function () {
        justDragged = true;
    });
    node.addEventListener('click', selectHandler(position, node, false));
}

function selectHandler(location, node, touch) {
    return function selectHandler(event) {
        // prevent dragging causing selection
        if (touch && (justDragged || new Date() - startTouch < 100)) {
            justDragged = false;
            return;
        }
        event.stopPropagation();
        event.preventDefault();
        if (event.handled === true) {
            return;
        }
        event.handled = true;
        // see if we're already selected
        var index = selected.indexOf(location);
        if (index >= 0) {
            selected.splice(index, 1);
            node.classList.remove('selected');
            node.classList.remove('animate-in');
        } else {
            selected.push(location);
            node.classList.add('selected');
        }
        if (selected.length === 3) {
            ws.send(JSON.stringify({Type: "play", Play: selected, Version: version}));
            // request animation frame
            setTimeout(function(){
                for (var s = 0; s < selected.length; s++) {
                    map[selected[s]].classList.remove('selected');
                }
                selected = [];
            }, 1);
        }
    }
}

var version = 0;
ws.onmessage = function (e) {
    var node;
    var data = JSON.parse(e.data);
    switch (data.Type) {
        case "cookie":
            document.cookie = data.Cookie;
            console.log("Set cookie", data.Cookie);
            break;
        case 'meta':
            gameId.textContent = data.GameId;
            updatePlayers(data.Players);
            location.hash = gameId.textContent;
            version = data.Version;
            break;
        case 'all':
            while (cards.firstChild) {
                cards.removeChild(cards.firstChild);
            }
            map = {};
            selected = [];
            // fallthrough
        case 'update':
            // remove all selections, someone else played
            for (var s = 0; s < selected.length; s++) {
                map[selected[s]].classList.remove('selected');
            }
            selected = [];
            location.hash = gameId.textContent;

            for (var i = 0; i < data.Updates.length; i++) {
                var update = data.Updates[i];
                // grab the node from the lookup map
                node = map[update.Location];
                if (node) {
                    // clear existing card, if any
                    while (node.firstChild) {
                        node.removeChild(node.firstChild);
                    }
                    node.classList.remove('animate-in');
                } else {
                    // create it if it doesn't exist
                    node = cards.appendChild(document.createElement('span'));
                    node.setAttribute('id', update.Location);
                    addEventListeners(update.Location, node);
                    map[update.Location] = node;
                }
                var shape;
                switch (update.Card.s) {
                    case 'p':
                        shape = pill;
                        break;
                    case 'd':
                        shape = diamond;
                        break;
                    case 'n':
                        shape = nut;
                        break;
                }
                for (var j = 0; j < update.Card.a; j++) {
                    node.appendChild(shape.cloneNode(true));
                }
                node.className = 'card ' + pattern(update.Card.p) + " " + color(update.Card.c);
                // request animation frame
                setTimeout(function(n) {
                    return function() { n.classList.add('animate-in'); }
                }(node), 1);
            }
            version = data.Version;
            gameId.textContent = data.GameId;
            updatePlayers(data.Players);
            break;
        default:
            console.log("unknown type", data);
    }
};

function pattern(p) {
    if (p === "h") {
        return "hollow";
    }
    if (p === "s") {
        return "solid";
    }
    if (p === "z") {
        return "striped";
    }
}

function color(c) {
    if (c === "r") {
        return "red"
    }
    if (c === "p") {
        return "purple"
    }
    if (c === "g") {
        return "green"
    }
}

ws.onerror = function (e) {
    console.log('error', e);
};

ws.onclose = function (e) {
    document.getElementById('disconnected').className = '';
};

var players = document.getElementById('players');
var cachedPlayers = [];
function updatePlayers(playerData) {
    if (cachedPlayers === playerData) {
        return;
    }
    while (players.firstChild) {
        players.removeChild(players.firstChild);
    }
    for(var i=0; i<playerData.length; i++) {
        var p = document.createElement('p');
        p.textContent = 'Player ' + playerData[i].Id + ': ' + playerData[i].Score;
        players.appendChild(p);
    }
}

document.getElementById('newgame').addEventListener('click', function(e) {
    if (confirm('Start a new game and abandon this one?')) {
        location.href = "";
    }
    e.stopPropagation();
    e.preventDefault();
});
