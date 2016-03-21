var giffyServices = angular.module('giffy.services', []);

giffyServices.service("localSession", function() {
    var localSession = {};

    var _toObject = function () {
        if (localStorage) {
            var the_obj = {};
            for (var x = 0; x < localStorage.length; x++) {
                var key = localStorage.key(x);
                the_obj[key] = localStorage.getItem(key);
            }
            return the_obj;
        } else {
            return localSession;
        }
    };

    return {
        get : function(key) {
            if(localStorage) {
                var jsonValue = localStorage.getItem(key);
                return JSON.parse(jsonValue);
            } else {
                return localSession[key];
            }
        },
        set : function(key, value) {
            if(localStorage) {
                var jsonValue = JSON.stringify(value);
                localStorage.setItem(key, jsonValue);
            } else {
                localSession[key] = value;
            }
        }, 
        has : function(key) {
            if(localStorage) {
                if(localStorage.getItem(key)){
                    return true;
                } else {
                    return false;
                }
            } else {
                if(key in localSession) {
                    return true;
                } else {
                    return false;
                }
            }
        },
        purge: function(key) {
            if(typeof(key) !== 'undefined') {
                if(localStorage) {
                    localStorage.removeItem(key);
                } else {
                    delete(localSession[key]);
                }
            }
            else {
                if(localStorage) {
                    localStorage.clear();
                } else {
                    delete localSession;
                    localSession = {};
                }
            }
        },
        serialize: function () {
            var theString = "";
            if (localStorage) {
                var contents = _toObject();
                theString = JSON.stringify(contents);
            } else {
                theString = JSON.stringify(localSession);
            }
            return theString;
        },
        deserialize: function (serialized) {
            var deserialized = JSON.parse(serialized);
            if (localStorage) {
                for (var key in deserialized) {
                    localStorage.setItem(key, deserialized[key]);
                }
            } else {
                localSession = deserialized;
            }
        },
        toObject : _toObject,
        toArray: function () {
            var kvps = [];
            if (localStorage) {
                for (var x = 0; x < localStorage.length; x++) {
                    var key = localStorage.key(x);
                    var value = localStorage.getItem(key);
                    kvps.push({ key: key, value: value });
                }
            } else {
                for (var key in localSession) {
                    var value = localSession[key];
                    kvps.push({ key: key, value: value });
                }
            }
        }
    };
});
