var giffyDirectives = angular.module('giffy.directives', []);

// --------------------------------------------------------------------------------
// VoteAPI
// --------------------------------------------------------------------------------

giffyDirectives.factory('voteAPI', ["$http",
	function ($http) {
		this.upvote = function (imageUUID, tagUUID) {
			return $http.post("/api/vote.up/" + imageUUID + "/" + tagUUID, null);
		};

		this.downvote = function (imageUUID, tagUUID) {
			return $http.post("/api/vote.down/" + imageUUID + "/" + tagUUID, null);
		};

		this.deleteUserVote = function (imageUUID, tagUUID) {
			return $http.delete("/api/user.vote/" + imageUUID + "/" + tagUUID);
		}

		this.deleteLink = function (imageUUID, tagUUID) {
			return $http.delete("/api/link/" + imageUUID + "/" + tagUUID);
		}

		return this;
	}
]);

// --------------------------------------------------------------------------------
// Header
// --------------------------------------------------------------------------------

giffyDirectives.directive("giffyHeader", function () {
	return {
		restrict: 'E',
		controller: "giffyHeaderController",
		templateUrl: "/static/partials/controls/header.html"
	}
});
giffyDirectives.controller('giffyHeaderController', ["$scope", "$http", function ($scope, $http) { }]);

// --------------------------------------------------------------------------------
// Footer
// --------------------------------------------------------------------------------

giffyDirectives.directive("giffyFooter", function () {
	return {
		restrict: 'E',
		controller: "giffyFooterController",
		templateUrl: "/static/partials/controls/footer.html"
	}
});
giffyDirectives.controller('giffyFooterController', ["$scope", function ($scope) { }]);

// --------------------------------------------------------------------------------
// Image
// --------------------------------------------------------------------------------

giffyDirectives.directive("giffyImage", function () {
	return {
		restrict: 'E',
		scope: {
			image: '='
		},
		controller: "giffyImageController",
		templateUrl: "/static/partials/controls/image.html"
	}
});
giffyDirectives.controller('giffyImageController', ["$scope",
	function ($scope) {
		var minH = 250;

		// portrait = width 100%
		$scope.isPortrait = function () {
			return false;
		};

		// landscape = height 100%
		$scope.isLandscape = function () {
			return !$scope.isPortrait();
		}
	}
]);

// --------------------------------------------------------------------------------
// Username
// --------------------------------------------------------------------------------

giffyDirectives.directive("userDetail", function () {
	return {
		restrict: 'E',
		scope: {
			user: '=',

		},
		controller: "UserDetailElementController",
		templateUrl: "/static/partials/controls/username.html"
	}
});
giffyDirectives.controller('UserDetailElementController', ["$scope", function ($scope) { }]);

// --------------------------------------------------------------------------------
// Search Box
// --------------------------------------------------------------------------------

giffyDirectives.directive("searchBox", function () {
	return {
		restrict: 'E',
		scope: {
			searchQuery: '='
		},
		controller: "SearchBoxController",
		templateUrl: "/static/partials/controls/search.html"
	}
});
giffyDirectives.controller('SearchBoxController', ["$scope", "$location",
	function ($scope, $location) {
		$scope.search = function (query) {
			if (query && query.length > 0) {
				$location.path("/search/" + query).replace();
			}
		};
	}]);


// --------------------------------------------------------------------------------
// Vote Button
// --------------------------------------------------------------------------------

giffyDirectives.directive('voteButton',
	function () {
		return {
			restrict: 'E',
			scope: {
				type: '=',
				link: '=',
				userVote: '=',
				object: '=',
				currentUser: '='
			},
			controller: 'voteButtonController',
			templateUrl: '/static/partials/controls/vote_button.html'
		};
	}
);
giffyDirectives.controller('voteButtonController', ["$scope", "voteAPI",
	function ($scope, voteAPI) {
		$scope.vote = function (isUpvote) {
			if (!$scope.hasVote()) {
				if (isUpvote) {
					voteAPI.upvote($scope.imageUUID(), $scope.tagUUID()).then($scope.onVote);
				} else {
					voteAPI.downvote($scope.imageUUID(), $scope.tagUUID()).then($scope.onVote);
				}
			} else {
				voteAPI.deleteUserVote($scope.imageUUID(), $scope.tagUUID()).then($scope.onVote);
			}
		};

		$scope.delete = function () {
			voteAPI.deleteLink($scope.imageUUID(), $scope.tagUUID()).then($scope.onVote);
		}

		$scope.isOnlyVoteCount = function () {
			if ($scope.type === "image") {
				if (!$scope.currentUser.is_logged_in) {
					return true;
				}
			}
			return false;
		}

		$scope.userIsLoggedIn = function () {
			return $scope.currentUser.is_logged_in && !$scope.currentUser.is_banned;
		}

		$scope.onVote = function (res) {
			$scope.$emit('voted');
		}

		$scope.tagUUID = function () {
			return $scope.link.tag_uuid;
		}

		$scope.imageUUID = function () {
			return $scope.link.image_uuid;
		}

		$scope.detailURL = function () {
			return "/tag/" + $scope.object.tag_value;
		}

		$scope.detailValue = function () {
			return $scope.object.tag_value;
		}

		$scope.canEdit = function () {
			return ($scope.currentUser.is_moderator || $scope.object.created_by == $scope.currentUser.uuid) && !$scope.currentUser.is_banned;
		}

		$scope.hasVote = function () {
			return !!$scope.userVote;
		};

		$scope.didUpvote = function () {
			return $scope.userVote && $scope.userVote.is_upvote;
		};

		$scope.didDownvote = function () {
			return $scope.userVote && !$scope.userVote.is_upvote;
		};
	}
]);

giffyDirectives.directive('ngEnter', function () {
	return function (scope, element, attrs) {
		element.bind("keydown keypress", function (event) {
			if (event.which === 13) {
				scope.$apply(function () {
					scope.$eval(attrs.ngEnter, { 'event': event });
				});

				event.preventDefault();
			}
		});
	};
});

giffyDirectives.directive('convertToNumber', function () {
	return {
		require: 'ngModel',
		link: function (scope, element, attrs, ngModel) {
			ngModel.$parsers.push(function (val) {
				return parseInt(val, 10);
			});
			ngModel.$formatters.push(function (val) {
				return '' + val;
			});
		}
	};
});