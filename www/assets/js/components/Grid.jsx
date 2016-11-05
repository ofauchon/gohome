var React = require('react');
var ReactTransitionGroup = require('react-addons-transition-group');
var ReactDOM = require('react-dom');
var ClassNames = require('classnames');

var ExpanderWrapper = React.createClass({
    // Part of the calls for TransitionGroup, have to set the CSS property
    // to the initial value, then after a small delay set the end animation value
    componentWillAppear: function(cb) {
        var $this = $(ReactDOM.findDOMNode(this)).find('.animateWrapper');
        $this.css({ 'margin-top': -500 });
        setTimeout(function() {
            cb();
        }, 10);
    },
    componentDidAppear: function() {
        var $this = $(ReactDOM.findDOMNode(this)).find('.animateWrapper');
        $this.css({ 'margin-top': '0px' });
    },
    componentWillLeave: function(cb) {
        var $this = $(ReactDOM.findDOMNode(this)).find('animateWrapper');
        $this.css({ 'margin-top': -500 });
        cb();
    },
    render: function() {
        return (
            <div className="cmp-ExpanderWrapper">
                <div className="animateWrapper">
                    {this.props.children}
                </div>
            </div>
        );
    }
});

var Grid = React.createClass({
    getDefaultProps: function() {
        return {
            cells: [],
            cellWidth: 110,
            cellHeight: 110,
            paddingLeft: 0,
            paddingRight: 0,
            paddingTop: 0,
            paddingBottom: 12,
            spacingH: 0,
            spacingV: 0
        };
    },

    getInitialState: function() {
        return {
            cellIndices: { x:-1, y:-1 },
            expanderIndex: -1,
            selectedIndex: -1,
            expanded: false,
            cellWidth: this.props.cellWidth,
            cellHeight: this.props.cellHeight
        }
    },

    componentWillReceiveProps: function(nextProps) {
        // If the cells changed, then we are goign to re-render, close the expander
        if (nextProps.cells && (nextProps.cells !== this.props.cells)) {
            this.closeExpander();
        }
    },
    
    shouldComponentUpdate: function(nextProps, nextState) {
        //TODO: Fix
        return true;
        if (nextProps.cells && (nextProps.cells != this.props.cells)) {
            return true;
        }
        if (nextState.cellWidth && (nextState.cellWidth !== this.state.cellWidth)) {
            return true;
        }
        if (nextState.cellHeight && (nextState.cellHeight !== this.state.cellHeight)) {
            return true;
        }
        if (nextState.expanderIndex != undefined && (nextState.expanderIndex !== this.state.expanderIndex)) {
            return true;
        }
        return false;
    },
    
    calcCellDimensions: function() {
        var $this = $(ReactDOM.findDOMNode(this));
        var gridWidthNoPadding = $this.width();
        var cellsPerRow = Math.floor(gridWidthNoPadding / this.props.cellWidth);

        var fittedWidth = Math.floor(this.props.cellWidth + (gridWidthNoPadding - (cellsPerRow * this.props.cellWidth)) / cellsPerRow);
        return {
            width: fittedWidth,
            height: fittedWidth
        };
    },

    componentDidMount: function() {
        var dimensions = this.calcCellDimensions();
        this.setState({
            cellWidth: dimensions.width,
            cellHeight: dimensions.height
        });
    },

    closeExpander: function() {
        this.setState({
            expanded: false,
            cellIndices: { x:-1, y:-1},
            expanderIndex: -1,
            selectedIndex: -1,
            expanderContent: null
        });
    },
    
    cellClicked: function(evt) {
        var $this = $(ReactDOM.findDOMNode(this));

        //width() returns without padding
        var gridWidthNoPadding = $this.width();
        
        var cellsPerRow = Math.floor(gridWidthNoPadding / this.state.cellWidth);

        var $target = $(evt.target);
        var targetPos = $target.position();

        var cellXPos = Math.floor((targetPos.left) / this.state.cellWidth);
        var yOffset = targetPos.top;

        var cellIndex = $target.data('cell-index')
        if (this.state.expanded) {
            // Have to take into account the expander height when calculating which
            // cell the user is clicking on
            if (cellIndex > (this.state.expanderIndex - 1)) {
                var expanderHeight = $this.find('.cmp-ExpanderWrapper').height();
                yOffset -= expanderHeight;
            }
        }

        var cellYPos = Math.floor(yOffset / this.props.cellHeight);
        var expanderIndex = Math.min(this.props.cells.length, (cellYPos + 1) * cellsPerRow);

        if (cellXPos === this.state.cellIndices.x &&
            cellYPos === this.state.cellIndices.y) {
            this.closeExpander();
        }
        else {
            this.setState({
                cellIndices: { x: cellXPos, y: cellYPos },
                expanderIndex: expanderIndex,
                selectedIndex: cellYPos * cellsPerRow + cellXPos,
                expanded: true,
                expanderContent: this.props.cells[cellIndex].content
            });
        }
    },

    expanderWillMount: function(content) {
        this.props.expanderWillMount && this.props.expanderWillMount(content);
    },
    
    render: function() {
        function makeCellWrapper(index, selectedIndex, cell) {
            var content = cell.cell;
            return (
                <div
                    key={cell.key}
                    ref={"cellWrapper-" + index}
                    onClick={this.cellClicked}
                    className="cellWrapper pull-left"
                    data-cell-index={index}
                    style={{
                        width: this.state.cellWidth,
                        height: this.state.cellHeight,
                        //marginRight: this.props.spacingH,
                        //marginBottom: this.props.spacingV
                    }}>
                    {content}
                    <i className={ClassNames({
                                 "fa": true,
                                 "fa-caret-up": true,
                                 "hidden": index !== selectedIndex})}></i>
                </div>
            );
        }

        var content = [];
        var key = '';
        if (this.state.selectedIndex !== -1) {
            key = this.props.cells[this.state.selectedIndex].key;
        }
        var transitionGroup = (
            <ReactTransitionGroup key={key + "transition"}>
                <ExpanderWrapper key={key + 'wrapper'}>
                    {this.state.expanderContent}
                </ExpanderWrapper>
            </ReactTransitionGroup>
        );

        for (var i=0; i<this.props.cells.length; ++i) {
            content.push(makeCellWrapper.bind(this)(i, this.state.selectedIndex, this.props.cells[i]));
            if (this.state.expanded && this.state.expanderIndex === (i + 1)) {
                content.push(transitionGroup);
            }
        }

        if (!this.state.expanded) {
            //TODO: This isn't working, looks like multiple transition groups are causing issues, revist.
            //This should make the expander animate closed, but seems to have some bugs
            //content.push(transitionGroup);
        }

        return (
            <div className="cmp-Grid" style={{
                paddingLeft: this.props.paddingLeft,
                paddingRight: this.props.paddingRight,
                paddingTop: this.props.paddingTop,
                paddingBottom: this.props.paddingBottom,
            }}>
                <div className="clearfix beforeExpander">
                    {content}
                    <div style={{clear:"both"}}></div>
                </div>
            </div>
        );
    }
});
module.exports = Grid;
