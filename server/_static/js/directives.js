var giffyDirectives = angular.module('giffy.directives', []);

giffyDirectives.factory('VoteAPI', function($http) {
  this.upvote = function() {
    return $http.post("/api/upvote/" + $routeParams.image_id + "/" + tagUUID, null);
  };

  this.downvote = function() {
    return $http.post("/api/downvote/" + $routeParams.image_id + "/" + tagUUID, null);
  };
});

giffyDirectives.directive('voteButton', function() {
  return {
    restrict: 'E',
    scope: {
      vote: '='
    },
    controller: 'VoteButtonCtrl'
  };
});

giffyDirectives.controller('VoteButtonCtrl', function($scope, VoteAPI) {
  $scope.vote = function(tagUUID, isUpvote) {
    if (!$scope.hasVote(tagUUID)) {
      if (isUpvote) {
        VoteAPI.upvote().success(fetchImageData);
      } else {
        VoteAPI.downvote().success(fetchImageData);
      }
    }
  };

  $scope.hasVote = function() {
    return !!$scope.vote;
  };

  $scope.didUpvote = function() {
    return $scope.vote && $scope.vote.is_upvote;
  };

  $scope.didDownvote = function() {
    return $scope.vote && $scope.vote.is_upvote;
  };
});
