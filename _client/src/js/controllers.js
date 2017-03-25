var giffyControllers = angular.module("giffy.controllers", []);

giffyControllers.controller('homeController', ["$scope", "$http", "$routeParams", "$location", "currentUser",
	function ($scope, $http, $routeParams, $location, currentUser) {
		currentUser($scope);

		$http.get("/api/images/random/12").then(function (res) {
			$scope.images = res.data.Response;
		});

		$http.get("/api/tags/random/24").then(function (res) {
			$scope.tags = res.data.Response;
		});
	}
]);

giffyControllers.controller("searchController", ["$scope", "$http", "$routeParams", "$location", "currentUser",
	function ($scope, $http, $routeParams, $location, currentUser) {
		currentUser($scope);

		$http.get("/api/images.search?query=" + $routeParams.search_query).then(function (res) {
			$scope.images = res.data.Response;
			$scope.searchQuery = $routeParams.search_query;
			$scope.searchedQuery = $routeParams.search_query;
		});

		$scope.searchImages = function (searchQuery) {
			if (searchQuery && searchQuery.length > 0) {
				$location.path("/search/" + searchQuery).replace();
			} else {
				$scope.images = [];
			}
		};
	}
]);

giffyControllers.controller("searchSlackController", ["$scope", "$http", "$routeParams", "$location", "currentUser",
	function ($scope, $http, $routeParams, $location, currentUser) {
		currentUser($scope);

		$http.get("/integrations/slack?text=" + $routeParams.search_query).then(function (res) {
			$scope.image_uuid = datums.image_uuid;
			$scope.image_url = datums.attachments[0].image_url;
			$scope.searchQuery = $routeParams.search_query;
			$scope.searchedQuery = $routeParams.search_query;
		});

		$scope.searchImages = function (searchQuery) {
			if (searchQuery && searchQuery.length > 0) {
				$location.path("/search.slack/" + searchQuery).replace();
			} else {
				$scope.images = [];
			}
		};
	}
]);


giffyControllers.controller("addImageController", ["$scope", "$http", "currentUser",
	function ($scope, $http, currentUser) {
		currentUser($scope, function () {
			if (!$scope.currentUser.is_logged_in) {
				window.location = "/";
			}
		});
	}
]);

giffyControllers.controller("imageController", ["$scope", "$http", "$routeParams", "$q", "currentUser",
	function ($scope, $http, $routeParams, $q, currentUser) {
		currentUser($scope, function () {
			fetchImageData();
			fetchImageStats();
		});

		$scope.addTags = function () {
			var tagValues = [];
			for (var x = 0; x < $scope.newTags.length; x++) {
				tagValues.push($scope.newTags[x].text);
			}

			$http.post("/api/tags/", tagValues).then(function (res) {
				var tags = res.data.Response;
				var calls = []
				angular.forEach(tags, function (tag) {
					calls.push($http.post("/api/vote.up/" + $scope.image.uuid + "/" + tag.uuid, {}));
				});

				$q.all(calls).then(function () {
					jQuery("#add-tag-modal").modal('hide');
					$scope.newTags = [];
					fetchTagData();
				});
			});
		};

		$scope.deleteImage = function () {
			if ($scope.currentUser.is_moderator) {
				if (confirm("are you sure?")) {
					$http.delete("/api/image/" + $scope.image.uuid).then(function () {
						window.location = "/";
					});
				}
			}
		};

		$scope.updateImageContentRating = function () {
			if ($scope.currentUser.is_moderator) {
				$http.put("/api/image/" + $scope.image.uuid, { content_rating: $scope.image.content_rating }).then(function (res) {
					$scope.image = res.data.Response;
				});
			}
		};

		var fetchImageData = function () {
			delete $scope.image;

			$http.get("/api/image/" + $routeParams.image_id).then(function (res) {
				$scope.image = res.data.Response;
				$scope.slackCommand = "/giffy img:" + $scope.image.uuid;
				fetchTagData();
			}, function (res) {
				window.location = "/";
			});
		};

		var fetchImageStats = function () {
			delete $scope.image_stats;
			$http.get("/api/image.stats/" + $routeParams.image_id).then(function (res) {
				$scope.image_stats = res.data.Response;
			}, function (res) {
				window.location = "/";
			});
		};

		var fetchTagData = function () {
			delete $scope.tags;
			delete $scope.linkLookup;
			delete $scope.userVoteLookup;

			$http.get("/api/image.tags/" + $routeParams.image_id).then(function (res) {
				$scope.tags = res.data.Response;
			});

			$http.get("/api/image.votes/" + $routeParams.image_id).then(function (res) {
				var linkLookup = {};
				for (var x = 0; x < res.data.Response.length; x++) {
					var link = res.data.Response[x];
					linkLookup[link.tag_uuid] = link;
				}
				$scope.linkLookup = linkLookup;
			}, function (res) { });

			if ($scope.currentUser.is_logged_in && !$scope.currentUser.is_banned) {
				$http.get("/api/user.votes.image/" + $routeParams.image_id).then(function (res) {
					var userVoteLookup = {};
					for (var x = 0; x < res.data.Response.length; x++) {
						var vote = res.data.Response[x];
						userVoteLookup[vote.tag_uuid] = vote;
					}
					$scope.userVoteLookup = userVoteLookup;
				}, function (res) { });
			}
		}

		$scope.formatFileSize = function (fileSizeBytes) {
			if (fileSizeBytes > (1 << 30)) {
				return (fileSizeBytes / (1 << 30)).toFixed(2) + " GB";
			}

			if (fileSizeBytes > (1 << 20)) {
				return (fileSizeBytes / (1 << 20)).toFixed(2) + " MB";
			}

			if (fileSizeBytes > (1 << 10)) {
				return (fileSizeBytes / 1024).toFixed(2) + " KB";
			}

			return fileSizeBytes + " bytes"
		};

		$scope.searchTags = function (query) {
			return $q(function (resolve, reject) {
				$http.get("/api/tags.search/?query=" + query).then(function (res) {
					var values = [];
					for (var x = 0; x < res.data.Response.length; x++) {
						var tag = res.data.Response[x];
						values.push({ text: tag.tag_value });
					}
					resolve(values);
				});
			});
		}

		if (!!$routeParams["show_add_tags"]) {
			jQuery("#add-tag-modal").modal();
		}

		$scope.$on("voted", function () {
			fetchTagData();
		});

		$scope.tagAddedHandler = function () {
			jQuery("#tagsInput .tags input").focus();
		};

		jQuery("#slack-command-link").on('click', function () {
			var slackLink = document.querySelector("#slack-command-link");
			copyElement(slackLink);
			return false;
		});

		jQuery('#add-tag-modal').on('shown.bs.modal', function () {
			jQuery('#add-tag-value').focus();
		});
	}
]);

giffyControllers.controller("tagController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		currentUser($scope);

		// tag information
		$http.get("/api/tag/" + $routeParams.tag_id).then(function (res) {
			$scope.tag = res.data.Response;
			fetchVoteData();
		});

		var fetchVoteData = function () {
			delete $scope.linkLookup;
			delete $scope.userVoteLookup;

			$http.get("/api/tag.images/" + $scope.tag.uuid).then(function (res) {
				$scope.images = res.data.Response;
			});

			$http.get("/api/tag.votes/" + $scope.tag.uuid).then(function (res) {
				var linkLookup = {};
				for (var x = 0; x < res.data.Response.length; x++) {
					var link = res.data.Response[x];
					linkLookup[link.image_uuid] = link;
				}
				$scope.linkLookup = linkLookup;
			}, function (res) { });

			if ($scope.currentUser.is_logged_in) {
				$http.get("/api/user.votes.tag/" + $scope.tag.uuid).then(function (res) {
					var userVoteLookup = {};
					for (var x = 0; x < res.data.Response.length; x++) {
						var vote = res.data.Response[x];
						userVoteLookup[vote.image_uuid] = vote;
					}
					$scope.userVoteLookup = userVoteLookup;
				}, function (res) { });
			}
		}

		$scope.$on("voted", function () {
			fetchVoteData();
		});
	}
]);

giffyControllers.controller("userController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		currentUser($scope, function () {
			$http.get("/api/user/" + $routeParams.user_id).then(function (res) {
				$scope.user = res.data.Response;

				$http.get("/api/user.images/" + $routeParams.user_id).then(function (res) {
					$scope.images = res.data.Response;
				});
			});
		});

		$scope.promote = function () {
			var user = $scope.user;
			user.is_moderator = !user.is_moderator;
			$http.put("/api/user/" + $routeParams.user_id, user).then(function (res) {
				$scope.user = res.data.Response;
			});
		};

		$scope.ban = function () {
			if (confirm("Are you sure?")) {
				var user = $scope.user;
				user.is_banned = !user.is_banned;
				$http.put("/api/user/" + $routeParams.user_id, user).then(function (res) {
					$scope.user = res.data.Response;
				});
			}
		};
	}
]);

giffyControllers.controller("moderationLogController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		currentUser($scope);

		var pageSize = 50;

		$http.get("/api/moderation.log/pages/" + pageSize + "/0").then(function (res) {
			$scope.page = 0;
			$scope.log = res.data.Response;
		});

		$scope.hasPreviousPage = function () {
			return $scope.page > 0;
		};

		$scope.hasNextPage = function () {
			return !!$scope.log && $scope.log.length >= pageSize;
		};

		$scope.nextPage = function () {
			if ($scope.hasNextPage()) {
				$scope.page = $scope.page + 1;
				$http.get("/api/moderation.log/pages/" + pageSize + "/" + ($scope.page * pageSize)).then(function (res) {
					$scope.log = res.data.Response;
				});
			}
		};

		$scope.previousPage = function () {
			if ($scope.page > 0) {
				$scope.page = $scope.page - 1;
				$http.get("/api/moderation.log/pages/" + pageSize + "/" + ($scope.page * pageSize)).then(function (res) {
					$scope.log = res.data.Response;
				});
			}
		};
	}
]);


giffyControllers.controller("searchHistoryController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		currentUser($scope, function () {
			if (!$scope.currentUser.is_admin) {
				window.location = "/"
			}

			$http.get("/api/search.history/pages/" + pageSize + "/0").then(function (res) {
				$scope.page = 0;
				$scope.history = res.data.Response;
			});
		});

		var pageSize = 50;

		$scope.hasPreviousPage = function () {
			return $scope.page > 0;
		};

		$scope.hasNextPage = function () {
			return !!$scope.history && $scope.history.length >= pageSize;
		};

		$scope.nextPage = function () {
			if ($scope.hasNextPage()) {
				$scope.page = $scope.page + 1;
				$http.get("/api/search.history/pages/" + pageSize + "/" + ($scope.page * pageSize)).then(function (res) {
					$scope.history = res.data.Response;
				});
			}
		};

		$scope.previousPage = function () {
			if ($scope.page > 0) {
				$scope.page = $scope.page - 1;
				$http.get("/api/search.history/pages/" + pageSize + "/" + ($scope.page * pageSize)).then(function (res) {
					$scope.history = res.data.Response;
				});
			}
		};
	}
]);

giffyControllers.controller("userSearchController", ["$scope", "$http", "currentUser",
	function ($scope, $http, currentUser) {
		currentUser($scope, function () {
			$http.get("/api/users/pages/50/0").then(function (res) {
				$scope.users = res.data.Response;
			});
		});

		$scope.searchUsers = function () {
			if ($scope.searchQuery) {
				$http.get("/api/users.search?query=" + $scope.searchQuery).then(function (res) {
					$scope.users = res.data.Response;
					$scope.searchedQuery = $scope.searchQuery;
				});
			} else {
				delete $scope.users;
			}
		};

		jQuery("#giffy-user-search-bar").focus();
	}
]);

giffyControllers.controller("logoutController", ["$scope", "$http",
	function ($scope, $http) {
		$http.post("/api/logout", null).then(function () {
			window.location = "/"
		});
	}
]);

giffyControllers.controller("slackCompleteController", ["$scope", "currentUser",
	function ($scope) {
		currentUser($scope);
	}
]);

giffyControllers.controller("aboutController", ["$scope", "$http", "currentUser",
	function ($scope, $http, currentUser) {
		currentUser($scope);

		$http.get("/api/images/random/1").then(function (res) {
			$scope.image = res.data.Response[0];
		});
	}
]);

giffyControllers.controller("notFoundController", function () { });

giffyControllers.controller("statsController", ["$scope", "$http", "currentUser",
	function ($scope, $http, currentUser) {
		currentUser($scope);

		$http.get("/api/stats").then(function (res) {
			$scope.stats = res.data.Response;
		});
	}
]);

giffyControllers.controller("teamsController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		var fetchTeams = function () {
			$http.get("/api/teams").then(function (res) {
				$scope.teams = res.data.Response;
			});
		}

		currentUser($scope, function () {
			if (!$scope.currentUser.is_admin) {
				window.location = "/"
			}

			fetchTeams();
		});

		$scope.updateTeamEnabled = function (team) {
			$http.put("/api/team/" + team.team_id, team).then(function () {
				fetchTeams();
			});
		};

		$scope.updateTeamContentRatingFilter = function (team) {
			$http.put("/api/team/" + team.team_id, team).then(function () {
				fetchTeams();
			});
		};
	}
]);

giffyControllers.controller("errorsController", ["$scope", "$http", "$routeParams", "currentUser",
	function ($scope, $http, $routeParams, currentUser) {
		var fetchErrors = function () {
			$http.get("/api/errors/50/0").then(function (res) {
				$scope.errors = res.data.Response;
			});
		}

		currentUser($scope, function () {
			if (!$scope.currentUser.is_admin) {
				window.location = "/"
			}

			fetchErrors();
		});
	}
]);