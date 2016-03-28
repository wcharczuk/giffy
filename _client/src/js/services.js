var giffyServices = angular.module('giffy.services', []);

giffyServices.service('giffyAuth', function() {

});

giffyServices.service("localStorage", function() {
    var _toObject = function () {
		var obj = {};
		for (var x = 0; x < localSession.length; x++) {
			var key = localSession.key(x);
			obj[key] = localSession.getItem(key);
		}
		return obj;
    };

    return {
        get : function(key) {
			var jsonValue = localStorage.getItem(key);
			return JSON.parse(jsonValue);
        },
        set : function(key, value) {
            var jsonValue = JSON.stringify(value);
			localStorage.setItem(key, jsonValue);
        },
        has : function(key) {
			if(localStorage.getItem(key)){
				return true;
			}
			return false;
        },
        purge: function(key) {
            if(!!key) {
				localStorage.removeItem(key);
			} else {
				localStorage.clear();
            }
        },
        serialize: function () {
			var contents = _toObject();
			return JSON.stringify(contents);
        },
        deserialize: function (serialized) {
            var deserialized = JSON.parse(serialized);
			for (var key in deserialized) {
				localStorage.setItem(key, deserialized[key]);
			}
        },
        toObject : _toObject,
        toArray: function () {
            var kvps = [];
			for (var x = 0; x < localStorage.length; x++) {
				var key = localStorage.key(x);
				var value = localStorage.getItem(key);
				kvps.push({ key: key, value: value });
			}
        }
    };
});
