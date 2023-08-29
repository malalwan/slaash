/////////////  /////
////////////  //////
///////////  ///////
//////////  ////////
/////////  /////////
////////  //////////
///////  ///////////
//////  ////////////
/////  /////////////

// Average Click Distance variables
var lastDetectedTime = new Date();
var averageClickDist = getCookie('cl_/_dist')
averageClickDist = averageClickDist ? averageClickDist : 0

document.addEventListener('DOMContentLoaded', function() {
  
    // PDP Button Variables
    var buttonSets = [
    ['.slider-button', '.quantity__button', '.button-show-more', '.product-popup-modal__toggle', '.product-popup-modal__button'],
    ['.button.button--full-width', '.thumbnail', '.drawer__close', '.quantity-popover__info-button', '.button.button--tertiary'],
    ['.button-close', '.cart__update-button', '.cart__checkout-button', '.product-form__submit', '.pickup-availability-button'],
    ['.quick-add__submit', '.quick-add-modal__toggle', '.button.button--tertiary.cart-remove-button', '.quantity-popover__info-button'],
    ['.cart__checkout-button', '.share-button__button']
    ];
    var buttons = [];
    buttonSets.forEach(function(set) {
        var setButtons = document.querySelectorAll(set.join(','));
        buttons = buttons.concat(Array.from(setButtons));
    });

    // Page Scroll Variables
    var scrollData = [];
    var recordingInterval = 5000; // 5 seconds in milliseconds
    var lastRecordedTime = 0;

    // Hover Action Variables
    var hoverData = [];
    var hoverSets = ['.grid__item.product__media-wrapper', '.product__info-wrapper'];
    var hovers = []
    hoverSets.forEach(function(set) {
        var elements = document.querySelectorAll(set);
        hovers = hovers.concat(Array.from(elements));
    });
    // Create an object to track the state of each hover element
    var hoverState = {};

    // PDP Button Actions
    buttons.forEach(function(button) {
        var buttonClass = button.getAttribute('class'); 
        button.addEventListener('click', function() {
            var timestamp = new Date();
            var eventData = {
                eventType: "buttonClick",
                timestamp: timestamp.toISOString(),
                eventData: buttonClass
            };
            sendDataToServer(eventData, 'event');
            var timeDiff = timestamp - lastDetectedTime;
            lastDetectedTime = timestamp;
            averageClickDist = (averageClickDist+timeDiff)/2;
            setCookie('cl_/_dist', averageClickDist, 365);
            var eventData = {
                eventType: "avgClickDist",
                timestamp: timestamp.toISOString(),
                eventData: averageClickDist
            };
            sendDataToServer(eventData, 'event');
        });
    });
  
    // Hover Actions
    hovers.forEach(function(hover) {
        var hoverTimeout;
        var isHovered = false;
        var hoverClass = hover.getAttribute('class');
    
        hover.addEventListener('mouseover', function() {
            clearTimeout(hoverTimeout);
    
            // Check if the element is in unhovered state before logging hover
            if (!isHovered) {
                hoverTimeout = setTimeout(function() {
                    var hoverClass = hover.getAttribute('class');
                    hoverData.push({
                        eventType: "hoverIn",
                        timestamp: new Date().toISOString(),
                        eventData: hoverClass
                    });
    
                    hoverState[hoverClass] = 'hovered';
                    isHovered = true;
    
                    var eventData = {
                        eventType: "hoverIn",
                        timestamp: new Date().toISOString(),
                        eventData: hoverClass
                    };
                    sendDataToServer(eventData, 'event');
                }, 300); // Delay before logging hover (adjust as needed)
            }
        });
    
        hover.addEventListener('mouseout', function() {
            clearTimeout(hoverTimeout);
    
            // Check if the element is in hovered state before logging unhover
            if (isHovered) {
                hoverTimeout = setTimeout(function() {
                    var hoverClass = hover.getAttribute('class');
                    hoverData.push({
                        eventType: "hoverOut",
                        timestamp: new Date().toISOString(),
                        eventData: hoverClass
                    });
    
                    hoverState[hoverClass] = 'unhovered';
                    isHovered = false;
    
                    var eventData = {
                        eventType: "hoverOut",
                        timestamp: new Date().toISOString(),
                        eventData: hoverClass
                    };
                    sendDataToServer(eventData, 'event');
                }, 300); // Delay before logging unhover (adjust as needed)
            }
        });
        hoverState[hoverClass] = 'unhovered';
    });

    // Scroll Depth Action
    window.addEventListener('scroll', function() {
        var currentTime = Date.now();

        // Check if the required interval has passed since the last recording
        if (currentTime - lastRecordedTime >= recordingInterval) {
            lastRecordedTime = currentTime;
            var scrollTop = window.pageYOffset || document.documentElement.scrollTop;
            var windowHeight = window.innerHeight || document.documentElement.clientHeight;
            var documentHeight = Math.max(
                document.body.scrollHeight,
                document.body.offsetHeight,
                document.documentElement.clientHeight,
                document.documentElement.scrollHeight,
                document.documentElement.offsetHeight
            );
            
            var scrollPercentage = (scrollTop / (documentHeight - windowHeight)) * 100;
            
            // Round to two decimal places for simplicity
            scrollPercentage = Math.round(scrollPercentage * 100) / 100;
    
            scrollData.push({
                scrollPercentage: scrollPercentage
            });
            var eventData = {
                eventType: "scrollDepth",
                timestamp: new Date().toISOString(),
                eventData: scrollPercentage.toString()
            };
            sendDataToServer(eventData, 'event');
        }
    });
  
    // avg time b/w clicks
  
});

// Utility Functions

function createAnonymousIdentifier(input) {
  const encoder = new TextEncoder();
  const data = encoder.encode(input);
  return crypto.subtle.digest('SHA-1', data).then(buffer => {
    const hashArray = Array.from(new Uint8Array(buffer));
    return hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('');
  });
}

function sendDataToServer(data, subdomain) {
    fetch('https://64bd-111-93-190-114.ngrok-free.app/'+subdomain, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        if (response.ok) {
            console.log('Event data sent successfully');
        } else {
            console.error('Failed to send event data');
        }
    })
    .catch(error => {
        console.error('Error sending event data:', error);
    });
}

// Function to get parameter by name from the URL
function getParameterByName(name, url = window.location.href) {
    name = name.replace(/[\[\]]/g, '\\$&');
    var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, ' '));
}

// Function to set a cookie
function setCookie(name, value, days) {
    var expires = "";
    if (days) {
        var date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "")  + expires + "; path=/";
}

// Function to get a cookie
function getCookie(name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for(var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}

// Create a session cookie
function setSessionCookie(name, value) {
    document.cookie = `${name}=${value}; path=/`;
}


// Variables
let source;
let pageCount;
let visitCount;
let host = window.location.hostname;
let path = window.location.pathname;
let browserLanguage = navigator.language || navigator.userLanguage;
let screenWidth = screen.width;
let screenHeight = screen.height;
let screenPixelDepth = screen.pixelDepth;
let screenColorDepth = screen.colorDepth;
let windowWidth = window.innerWidth || document.documentElement.clientWidth;
let windowHeight = window.innerHeight || document.documentElement.clientHeight;
let timezoneOffset = new Date().getTimezoneOffset();
let platform = navigator.platform;
let cookiesEnabled = navigator.cookieEnabled;
let supportsTouch = 'ontouchstart' in window;
let prefersDarkScheme = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
// Send device data to your local server
var data = {
    browserLanguage: browserLanguage,
    screenWidth: screenWidth,
    screenHeight: screenHeight,
    screenPixelDepth: screenPixelDepth,
    screenColorDepth: screenColorDepth,
    windowWidth: windowWidth,
    windowHeight: windowHeight,
    timezoneOffset: timezoneOffset,
    platform: platform,
    cookiesEnabled: cookiesEnabled,
    supportsTouch: supportsTouch,
    prefersDarkScheme: prefersDarkScheme
};

console.log(data);
var anonymousId = getCookie('an_##_/');

if (anonymousId == null) {
    let id;
    sendDataToServer(data, 'device');
    const date = new Date().toISOString();
    createAnonymousIdentifier(date).then(sha1Hash => {
        console.log('Date:', date);
        console.log('SHA-1 Hash:', sha1Hash);
        id = host+sha1Hash;
        setCookie('an_##_/', id, 365);
        anonymousId = id;
    });
}

function setRecurCookiesAndSend() {
  // Get page count from cookies, increment it and store it back
  // Get visit count from cookies, increment it if the page count DNE
  pageCount = getCookie('pa_#_ge');
  visitCount = getCookie('vi_#_sit');
  source = getParameterByName('utm_source') || 'direct';
  pageCount = pageCount ? parseInt(pageCount) + 1 : 1;
  visitCount = visitCount ? (pageCount ? parseInt(visitCount) : parseInt(visitCount)+1) : 1;
  setSessionCookie('pa_#_ge', pageCount);
  setCookie('vi_#_sit', visitCount, 365);
  setSessionCookie('sou_utm_rce', source);
  anonymousId = getCookie('an_##_/');
    var data = {
    anonymousId: anonymousId,
    source: source,
    pageCount: pageCount,
    visitCount: visitCount,
    host: host,
    path: path
  };

  console.log(data);

  sendDataToServer(data, 'essential');

}

setRecurCookiesAndSend();
