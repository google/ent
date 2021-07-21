//
// Copyright 2021 The Ent Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

class Item extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            item: props.item,
        };
    }

    render() {
        return e('div', { className: 'border py-3 my-3 bg-' + env.colour + '-200 hover:bg-' + env.colour + '-300' },
            e('span', { className: 'px-5 w-24 inline-block' }, '[' + this.state.item.score + ']'),
            e('a', { href: this.state.item.url, className: 'underline' }, this.state.item.title),
            e('span', { className: "p-2" }, 'by'),
            e('span', null, this.state.item.by),
        );
    }
}
