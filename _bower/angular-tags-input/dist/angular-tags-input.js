/**
 * @author Umed Khudoiberdiev <info@zar.tj>
 */
(function() {
    'use strict';

    /**
     * @ngdoc module
     * @name tagsInput
     * @description
     * A special angular directive that allows to add "tags" into input box.
     */
    angular.module('tagsInput', ['autoGrowInput', 'selectOptions', 'disableAll', 'ngSanitize']);

})();
angular.module("tagsInput").run(["$templateCache", function($templateCache) {$templateCache.put("tags-input.html","<div class=\"tags-input\" ng-attr-style=\"{{ getContainerWidth() > 0 ? \'max-width: \' + getContainerWidth()  + \'px\' : \'\' }}\" disable-all=\"disabled\">\n    <div ng-repeat=\"tag in getTags()\" class=\"token-item\">\n        <div class=\"token\" tabindex=\"{{ $index }}\"\n             ng-click=\"tagSelect($event, tag)\"\n             ng-class=\"{ selected: isSelected(tag) }\">\n            <div class=\"token-template\" ng-bind-html=\"getItemName(tag)\"></div>\n            <a class=\"token-remove\" ng-show=\"isRemoveButtonEnabled\" ng-click=\"tagRemove($event, $index)\">×</a>\n        </div>\n\n    </div>\n    <div class=\"token-item\">\n        <input auto-grow-input\n               ng-model=\"tokenInputValue\"\n               ng-change=\"onInputValueChange(tokenInputValue)\"\n               ng-attr-placeholder=\"{{ !getTags().length ? placeholder : \'\' }}\"\n               autocomplete=\"off\"\n               ng-disabled=\"isDisabled\">\n    </div>\n    <div class=\"loading\" ng-show=\"loadingInProgress\"></div>\n</div>\n");}]);
/**
 * @author Umed Khudoiberdiev <info@zar.tj>
 */
(function() {
    'use strict';

    /**
     * @ngdoc directive
     * @name tagsInput
     * @restrict E
     * @description
     * A special angular directive that allows to add "tags" into input box.
     *
     * @param {expression} ngModel Model that will be changed
     * @param {string} placeholder Text that will be added to input as default text (placeholder)
     * @param {number} minLength Minimal length of the tag to be added
     * @param {number} maxItems Maximal length of the tag to be added
     * @param {number} maxItems Maximal number of items that are allowed to be added to the tags
     * @param {boolean} uniqueNames Indicates if user can add only tags with unique names
     * @param {string} delimiters Array of characters, on press on which new tag will be added
     * @param {boolean} noPersist If set to true then user cannot add a new tag
     * @param {boolean} isRemoveButton Indicates if remove button is enabled
     * @param {boolean} isRestoreOnBackspace If set to true then tag will be restored (not deleted) to text when user press backspace
     * @param {Function} decorator Function used to decorate and change representation (view) of the tag
     * @param {number} caretPosition Used to control caret position in the tags input
     * @param {Function} tokenInputValue Text value inside input
     * @param {Function} containerWidth Component's width
     */
    angular.module('tagsInput').directive('tagsInput', tagsInput);

    /**
     * @ngdoc controller
     * @name TagsInputCtrl
     */
    angular.module('tagsInput').controller('TagsInputCtrl', TagsInputCtrl);

    /**
     * @ngInject
     */
    function tagsInput() {
        return {
            scope: {
                ngModel: '=',
                placeholder: '@',
                minLength: '=?',
                maxLength: '=?',
                maxItems: '=?',
                uniqueNames: '=?',
                delimiters: '@',
                disabled: '=?',
                noPersist: '=?',
                isRemoveButton: '=?',
                isRestoreOnBackspace: '=?',
                decorator: '=?',
                caretPosition: '=?',
                tokenInputValue: '=?',
                containerWidth: '=?'
            },
            replace: true,
            restrict: 'E',
            require: ['ngModel', 'selectOptions'],
            templateUrl: 'tags-input.html',
            controller: 'TagsInputCtrl'
        }
    }

    /**
     * @ngInject
     */
    function TagsInputCtrl($scope, $element, $attrs, SelectedTokensRegistry, ArrayNgModelHelper, SelectedTokensValidator, $timeout) {

        // ---------------------------------------------------------------------
        // Local variables
        // ---------------------------------------------------------------------

        var _this                   = this;
        var ngModelCtrl             = $element.controller('ngModel');
        var selectOptionsCtrl       = $element.controller('selectOptions');
        var selectedTokens          = new SelectedTokensRegistry();
        var ngModelHelper           = new ArrayNgModelHelper(ngModelCtrl);
        var selectedTokensValidator = new SelectedTokensValidator();
        var input                   = $element[0].querySelector('input');
        var nameField               = selectOptionsCtrl.getItemNameWithoutPrefixes();

        // ---------------------------------------------------------------------
        // Scope variables
        // ---------------------------------------------------------------------

        $scope.isRemoveButtonEnabled = true;

        // ---------------------------------------------------------------------
        // Configuration
        // ---------------------------------------------------------------------

        selectedTokensValidator.set({
            minLength:      $scope.minLength,
            maxLength:      $scope.maxLength,
            maxItems:       $scope.maxItems,
            nameField:      nameField,
            checkUnique:    $scope.uniqueNames
        });

        // ---------------------------------------------------------------------
        // Local functions
        // ---------------------------------------------------------------------

        /**
         * Changes a caret position and moves input to a new position.
         *
         * @param {number} position
         */
        var changeCaretPosition = function (position) {
            $scope.caretPosition = position;
            moveInputToPosition(position);
        };

        /**
         * Moves input to the given position in the tokens container.
         *
         * @param {int} caretPosition
         */
        var moveInputToPosition = function(caretPosition) {
            var containerElements = $element[0].getElementsByClassName('token-item');
            if (!containerElements[caretPosition]) return;
            $element[0].insertBefore(input, containerElements[caretPosition]);
        };

        /**
         * Clears tags input value.
         */
        this.clearInputValue = function () {
            $scope.tokenInputValue = ''; // restore input value
            input.value = '';
            input.dispatchEvent(new CustomEvent('update'));
            input.focus();
            dispatchTextEntered('');
        };

        var textEnteredCallbacks = [];
        /**
         * Registers a callback that will run when text is changed.
         * This is a replacement for tags-input.text_entered event.
         */
        this.onTextEntered = function(callback) {
            textEnteredCallbacks.push(callback);
        };

        var dispatchTextEntered = function(text) {
            angular.forEach(textEnteredCallbacks, function(callback) {
                callback(text);
            });
        };

        /**
         * Adds a new item to the model at the current cursor position. New item's value
         * got from the input. After adding a new item, input is cleared.
         */
        var addNewValueFromInput = function() {
            if (!$scope.tokenInputValue) return;
            var itemToBeAdded = selectOptionsCtrl.createItem($scope.tokenInputValue.trim());
            if ($scope.noPersist === true) return;
            if (selectedTokensValidator.canItemBeAdded(ngModelHelper.getAll(), itemToBeAdded) === false) return;

            ngModelHelper.add(itemToBeAdded, $scope.caretPosition);
            modelChanged();
            changeCaretPosition($scope.caretPosition + 1);
            _this.clearInputValue();
        };

        /**
         * Move input to the left in the tokens container.
         */
        var moveInputToLeft = function() {
            if ($scope.tokenInputValue || $scope.caretPosition <= 0) return;

            changeCaretPosition($scope.caretPosition - 1);
            input.focus();
        };

        /**
         * Move input to the right in the tokens container.
         */
        var moveInputToRight = function() {
            if ($scope.tokenInputValue || $scope.caretPosition >= ngModelHelper.count()) return;

            changeCaretPosition($scope.caretPosition + 1);
            input.focus();
        };

        /**
         * Removes all selected tokens from the model.
         */
        var removeSelectedTokens = function() {
            if ($scope.tokenInputValue || ngModelHelper.count() === 0) return;

            ngModelHelper.removeAll(selectedTokens.getAll());
            modelChanged();
            input.dispatchEvent(new CustomEvent('update'));
            changeCaretPosition(ngModelHelper.count());
            selectedTokens.clear();
            input.focus();
        };

        /**
         * Removes previous to the current caret position token from the model.
         */
        var removePreviousToken = function() {
            if ($scope.tokenInputValue || ngModelHelper.count() === 0 || $scope.caretPosition <= 0) return;

            changeCaretPosition($scope.caretPosition - 1);

            if ($scope.isRestoreOnBackspace) {
                var removedToken = ngModelHelper.get($scope.caretPosition);
                $scope.tokenInputValue = (nameField ? removedToken[nameField] : removedToken) + ' ';
                input.value = $scope.tokenInputValue;
                dispatchTextEntered(input.value);
            }

            ngModelHelper.removeAt($scope.caretPosition);
            modelChanged();
            moveInputToPosition($scope.caretPosition);
            input.dispatchEvent(new CustomEvent('update'));
            input.focus();
        };

        /**
         * Removes the next to the current caret position token from the model.
         */
        var removeNextToken = function() {
            if ($scope.tokenInputValue || ngModelHelper.count() === 0) return;

            ngModelHelper.removeAt($scope.caretPosition);
            modelChanged();
            moveInputToPosition($scope.caretPosition);
            input.dispatchEvent(new CustomEvent('update'));
            dispatchTextEntered(input.value);
            input.focus();
        };

        var modelChanged = function() {
            if ($attrs.onChange) {
                $timeout(function() {
                    selectOptionsCtrl.applyOnScope($attrs.onChange);
                });
            }
        };

        // ---------------------------------------------------------------------
        // Scope functions
        // ---------------------------------------------------------------------

        /**
         * Gets container width.
         *
         * @returns {number}
         */
        $scope.getContainerWidth = function() {
            return $scope.containerWidth ? $scope.containerWidth : $element[0].offsetWidth;
        };

        /**
         * Gets the items that will be displayed as tags.
         *
         * @returns {Object[]}
         */
        $scope.getTags = function() {
            var tags = ngModelCtrl.$modelValue;
            if (tags && !angular.isArray(tags))
                return [tags];

            return tags;
        };

        /**
         * Gets the item name that will be used to display in the list.
         *
         * @param {Object} item
         * @returns {string}
         */
        $scope.getItemName = function(item) {
            var value = selectOptionsCtrl.parseItemValueFromSelection(item);
            value = String(value).replace(/<[^>]+>/gm, ''); // strip html from the data here
            return $scope.decorator ? $scope.decorator(item) : value;
        };

        /**
         * Called when input value changes.
         */
        $scope.onInputValueChange = function(value) {
            input.dispatchEvent(new CustomEvent('update'));
            dispatchTextEntered(value);
            selectedTokens.clear(); // remove bulk selected elements if user started to type
        };

        /**
         * Checks if given token is in the list of selected tokens.
         *
         * @param {object} token
         * @returns {boolean} True if token is selected, false otherwise
         */
        $scope.isSelected = function(token) {
            return selectedTokens.has(token);
        };

        /**
         * When click is performed on a token - we "select" it, and remove
         * select effect from all other tokens if shift key was not pressed.
         *
         * @param {MouseEvent} event
         * @param {object} selectedToken
         */
        $scope.tagSelect = function(event, selectedToken) {
            event.stopPropagation();

            if (!event.altKey && !event.ctrlKey && !event.shiftKey) {
                selectedTokens.clear();
                selectedTokens.add(selectedToken);
            } else {
                selectedTokens.toggle(selectedToken);
            }

            changeCaretPosition(ngModelHelper.count());
        };

        /**
         * Removes a tag from the ng-model by a given tag index.
         *
         * @param {MouseEvent} event
         * @param {int} index
         */
        $scope.tagRemove = function(event, index) {
            event.stopPropagation();
            ngModelHelper.removeAt(index);
            modelChanged();
            changeCaretPosition(ngModelHelper.count());
            selectedTokens.clear();
            input.blur();
        };

        // ---------------------------------------------------------------------
        // DOM manipulation
        // ---------------------------------------------------------------------

        // when user clicks on the container it must automatically activate its input
        $element[0].addEventListener('click', function () {
            changeCaretPosition(ngModelHelper.count());
            input.focus();
            selectedTokens.clear(); // remove bulk selected elements if user started to type
            $scope.$digest();
        });

        // listen to key downs on the container to make control operations
        $element[0].addEventListener('keydown', function (e) {
            switch (e.keyCode) {

                case 65: // KEY "A"
                    if ((e.ctrlKey || e.metaKey) && !$scope.tokenInputValue) { // select all tokens
                        selectedTokens.addAll(ngModelHelper.getAll());
                        $scope.$digest();
                    }
                    return;

                case 13: // KEY "ENTER"
                    addNewValueFromInput();
                    $scope.$apply(); // note: don't use $digest here because it will cause a lag
                    return;

                case 37: // KEY "LEFT"
                    moveInputToLeft();
                    $scope.$apply(); // note: don't use $digest here because it will break proper caret position changing
                    return;

                case 39: // KEY "RIGHT"
                    moveInputToRight();
                    $scope.$apply(); // note: don't use $digest here because it will break proper caret position changing
                    return;

                case 8: // KEY "BACKSPACE"
                    selectedTokens.hasAny() ? removeSelectedTokens() : removePreviousToken();
                    $scope.$digest();
                    return;

                case 46: // KEY "DELETE"
                    selectedTokens.hasAny() ? removeSelectedTokens() : removeNextToken();
                    $scope.$digest();
                    return;
            }
        });

        // listen to key press to check if delimiter is pressed and we need to add a new tag
        $element[0].addEventListener('keypress', function (e) {
            var isAdded = false;
            angular.forEach($scope.delimiters, function(delimiter) {
                if (isAdded === false && e.keyCode === delimiter.charCodeAt(0)) {
                    addNewValueFromInput();
                    e.preventDefault();
                    $scope.$apply(); // note: don't use $digest here because it will cause a lag
                    isAdded = true;
                }
            });
        });

        // close dropdown, reset caret and other things if use clicks outside of this directive
        document.addEventListener('mousedown', function(event) {
            if ($element[0].contains(event.target)) return;

            selectedTokens.clear();
            changeCaretPosition(ngModelHelper.count());
            $scope.$digest();
        });

        // ---------------------------------------------------------------------
        // Watchers
        // ---------------------------------------------------------------------

        // this is required to set carent position to the end when ng model firstly initialized
        $scope.$watch('ngModel', function() {
            changeCaretPosition(ngModelHelper.count());
        });

        // watch a caret position changes from outside
        $scope.$watch('caretPosition', function(newPosition, oldPosition) {
            if (newPosition === oldPosition)
                return;
            changeCaretPosition(newPosition);
        });

        // listen for remove button to enable or disable it
        $scope.$watch('isRemoveButton', function(isRemoveButton) {
            if (!angular.isDefined($scope.isRemoveButton)) return;
            $scope.isRemoveButtonEnabled = isRemoveButton;
        });

        // ---------------------------------------------------------------------
        // Event listeners
        // ---------------------------------------------------------------------

        // this event request tags input to clear itself
        $scope.$on('tags-input.clear_input', function() {
            _this.clearInputValue();
        });

    }

})();
/**
 * @author Umed Khudoiberdiev <info@zar.tj>
 */
(function() {
    'use strict';

    /**
     * @ngdoc service
     * @name SelectedTokensValidator
     */
    angular.module('tagsInput').factory('SelectedTokensValidator', SelectedTokensValidator);

    /**
     * @constructor
     */
    function SelectedTokensValidator() {
        return function() {

            /**
             * Validation options.
             *
             * @type {{ nameField: string, minLength: number, maxLength: number, maxItems: number }}
             */
            var options;

            var getRealItemName = function(item) {
                return options.nameField ? item[options.nameField] : item;
            };

            /**
             * Sets the options required to perform validation.
             *
             * @param {object} validationOptions
             */
            this.set = function(validationOptions) {
                options = validationOptions;
            };

            /**
             * Determines if value is not unique and cannot be added to
             * the model array.
             *
             * @param {Array.<object>} tokens
             * @param {object} item
             * @returns {boolean}
             */
            this.isUnique = function(tokens, item) {
                if (!options.checkUnique) return true;

                var found = false;
                angular.forEach(tokens, function(token) {
                    var tokenName = getRealItemName(token);
                    var itemName  = getRealItemName(item);
                    if (tokenName && itemName && tokenName.toLowerCase() === itemName.toLowerCase())
                        found = true;
                });

                return found === false;
            };

            /**
             * Checks if item value is not empty.
             *
             * @param {boolean} item
             * @returns {boolean}
             */
            this.isNotEmpty = function(item) {
                return getRealItemName(item).length > 0;
            };

            /**
             * Checks if item value is not too short.
             *
             * @param {boolean} item
             * @returns {boolean}
             */
            this.isNotShort = function(item) {
                return angular.isUndefined(options.minLength) || getRealItemName(item).length >= options.minLength;
            };

            /**
             * Checks if item value is not too long.
             *
             * @param {boolean} item
             * @returns {boolean}
             */
            this.isNotLong = function(item) {
                return angular.isUndefined(options.maxLength) || getRealItemName(item).length <= options.maxLength;
            };

            /**
             * Checks if user didn't reach a maximal number of items in the input.
             *
             * @param {number} numberOfTokens
             * @returns {boolean}
             */
            this.isNotMaxAllowed = function(numberOfTokens) {
                return angular.isUndefined(options.maxItems) || numberOfTokens < options.maxItems;
            };

            /**
             * Determines if value can be added to the model array or not.
             *
             * @param {Array.<object>} tokens
             * @param {object} item
             * @returns {boolean}
             */
            this.canItemBeAdded = function(tokens, item) {
                return  this.isNotEmpty(item) &&
                    this.isUnique(tokens, item) &&
                    this.isNotShort(item) &&
                    this.isNotLong(item) &&
                    this.isNotMaxAllowed(tokens.length);
            };

        };
    }

})();
/**
 * @author Umed Khudoiberdiev <info@zar.tj>
 */
(function() {
    'use strict';

    /**
     * @ngdoc service
     * @name SelectedTokensRegistry
     */
    angular.module('tagsInput').factory('SelectedTokensRegistry', SelectedTokensRegistry);

    /**
     * @constructor
     */
    function SelectedTokensRegistry() {
        return function() {
            var selectedTokens = [];

            this.getAll = function() {
                return selectedTokens;
            };

            this.hasAny = function() {
                return selectedTokens.length !== 0;
            };

            this.has = function(token) {
                return selectedTokens.indexOf(token) !== -1;
            };

            this.add = function(token) {
                selectedTokens.push(token);
            };

            this.addAll = function(tokens) {
                angular.forEach(tokens, function(token) {
                    selectedTokens.push(token);
                });
            };

            this.remove = function(token) {
                var index = selectedTokens.indexOf(token);
                if (index !== -1)
                    selectedTokens.splice(index, 1);
            };

            this.toggle = function(token) {
                if (this.has(token)) {
                    this.remove(token);
                } else {
                    this.add(token);
                }
            };

            this.clear = function() {
                selectedTokens = [];
            };

        };
    }

})();
/**
 * @author Umed Khudoiberdiev <info@zar.tj>
 */
(function() {
    'use strict';

    /**
     * @ngdoc service
     * @name ArrayNgModelHelper
     */
    angular.module('tagsInput').factory('ArrayNgModelHelper', ArrayNgModelHelper);

    /**
     * @constructor
     */
    function ArrayNgModelHelper() {
        return function(ngModelCtrl) {
            var self = this;

            self.onChange = function(scope, callback) {
                scope.$watch(function () {
                    return ngModelCtrl.$modelValue;
                }, callback);

            };

            self.getAll = function() {
                return ngModelCtrl.$modelValue || [];
            };

            self.get = function(index) {
                var model = self.getAll();
                return model[index];
            };

            self.has = function(item) {
                var model = self.getAll();
                return model.indexOf(item) !== -1;
            };

            self.set = function(model) {
                ngModelCtrl.$setViewValue(model);
            };

            self.add = function(item, index) {
                var model = self.getAll();
                model.splice(index, 0, item);
                ngModelCtrl.$setViewValue(model);
            };

            self.remove = function(item) {
                var model = self.getAll();
                var index = model.indexOf(item);
                if (index !== -1)
                    model.splice(index, 1);
                ngModelCtrl.$setViewValue(model);
            };

            self.removeAt = function(index) {
                var model = self.getAll();
                model.splice(index, 1);
                ngModelCtrl.$setViewValue(model);
            };

            self.removeAll = function(items) {
                var model = self.getAll();

                angular.forEach(items, function(item) {
                    var index = model.indexOf(item);
                    if (index !== -1)
                        model.splice(index, 1);
                });

                ngModelCtrl.$setViewValue(model);
            };

            self.count = function() {
                return self.getAll().length;
            };

        };
    }

})();