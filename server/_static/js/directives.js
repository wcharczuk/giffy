var giffyDirectives = angular.module('giffy.directives', []);

giffyDirectives.factory('voteAPI', function($http) {
  this.upvote = function(imageUUID, tagUUID) {
    return $http.post("/api/vote.up/" + imageUUID + "/" + tagUUID, null);
  };

  this.downvote = function(imageUUID, tagUUID) {
    return $http.post("/api/vote.down/" + imageUUID + "/" + tagUUID, null);
  };
  
  this.deleteUserVote = function(imageUUID, tagUUID) {
    return $http.delete("/api/user.vote/" + imageUUID + "/" + tagUUID);
  }
  
  this.deleteLink = function(imageUUID, tagUUID) {
    return $http.delete("/api/link/" + imageUUID + "/" + tagUUID);
  }
  return this;
});


giffyDirectives.directive("giffyImage", function() {
  return {
    restrict: 'E',
    scope: {
      image: '=',
      user: '='
    },
    controller: "GiffyImageController",
    templateUrl: "/static/partials/image_element.html"
  }
});

giffyDirectives.controller('GiffyImageController', ["$scope", 
  function($scope) {
    $scope.deleteImage = function() {
      if (confirm("Are you sure?")) {
        $http.delete("/api/image/" + $routeParams.image_id).success(function(res) {
          $scope.$emit('image.deleted');
        });
      }
    }
  }
]);


giffyDirectives.directive("userDetail", function() {
  return {
    restrict: 'E',
    scope: {
      user: '='
    },
    controller: "UserDetailElementController",
    templateUrl: "/static/partials/user_detail_element.html"
  }
});

giffyDirectives.controller('UserDetailElementController', ["$scope", 
  function($scope) {}
]);

giffyDirectives.directive('voteButton',
  function() {
    return {
      restrict: 'E',
      scope: {
        type: '=',
        user: '=',
        link: '=',
        userVote: '=',
        object: '='
      },
      controller: 'VoteButtonController',
      templateUrl: '/static/partials/vote_button.html'
    };
  }
);

giffyDirectives.controller('VoteButtonController', ["$scope", "voteAPI", function($scope, voteAPI) {
    $scope.vote = function(isUpvote) {
      if (!$scope.hasVote()) {
        if (isUpvote) {
          voteAPI.upvote($scope.imageUUID(), $scope.tagUUID()).success($scope.onVote);
        } else {
          voteAPI.downvote($scope.imageUUID(), $scope.tagUUID()).success($scope.onVote);
        }
      } else {
        voteAPI.deleteUserVote($scope.imageUUID(), $scope.tagUUID()).success($scope.onVote);
      }
    };
    
    $scope.delete = function() {
      voteAPI.deleteLink($scope.imageUUID(), $scope.tagUUID()).success($scope.onVote);
    }

    $scope.isOnlyVoteCount = function() {
      if ($scope.type === "tag") {
        if (!$scope.user.is_logged_in) {
          return true;
        }
      }
      return false;
    }
    
    $scope.userIsLoggedIn = function() {
      return $scope.user.is_logged_in;
    }
    
    $scope.onVote = function(res) {
      $scope.$emit('voted');
    }

    $scope.tagUUID = function() {
      return $scope.link.tag_uuid;
    }
    
    $scope.imageUUID = function() {
      return $scope.link.image_uuid;
    }
    
    $scope.detailURL = function() {
      return "/#/tag/" + $scope.object.tag_value;
    }
    
    $scope.detailValue = function() {
      return $scope.object.tag_value;
    }

    $scope.canEdit = function() {
      return $scope.user.is_moderator || $scope.object.created_by == $scope.user.uuid;
    }

    $scope.hasVote = function() {
      return !!$scope.userVote;
    };

    $scope.didUpvote = function() {
      return $scope.userVote && $scope.userVote.is_upvote;
    };

    $scope.didDownvote = function() {
      return $scope.userVote && !$scope.userVote.is_upvote;
    };
  }
]);

giffyDirectives.directive('ngEnter', function() {
    return function(scope, element, attrs) {
        element.bind("keydown keypress", function(event) {
            if(event.which === 13) {
                scope.$apply(function(){
                    scope.$eval(attrs.ngEnter, {'event': event});
                });

                event.preventDefault();
            }
        });
    };
});