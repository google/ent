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

class Items extends React.Component {
    constructor() {
        super();
        this.state = {
            items: [
                { id: 0, by: "user", score: 126, title: "Pokegb: A gameboy emulator that only plays Pokémon Blue, in 68 lines of C++", url: "https://binji.github.io/posts/pokegb/" },
                { id: 1, by: "user", score: 126, title: "Pokegb: A gameboy emulator that only plays Pokémon Blue, in 68 lines of C++", url: "https://binji.github.io/posts/pokegb/" },
                { id: 2, by: "user", score: 126, title: "Pokegb: A gameboy emulator that only plays Pokémon Blue, in 68 lines of C++", url: "https://binji.github.io/posts/pokegb/" },
                { id: 3, by: "user", score: 126, title: "Pokegb: A gameboy emulator that only plays Pokémon Blue, in 68 lines of C++", url: "https://binji.github.io/posts/pokegb/" },
                { id: 4, by: "user", score: 126, title: "Pokegb: A gameboy emulator that only plays Pokémon Blue, in 68 lines of C++", url: "https://binji.github.io/posts/pokegb/" },
            ]
        };
    }

    render() {
        return e('div', null,
            this.state.items.map((item) => e(Item, { item: item, key: item.id }, null)));
    }

    componentDidMount() {
        this.go();
    }

    async go() {
        const topStories = await (await fetch("https://hacker-news.firebaseio.com/v0/topstories.json")).json()
        console.log(topStories);
        var items = [];
        for (var i = 0; i < env.numberOfArticles; i++) {
            const itemId = topStories[i];
            const item = await (await fetch("https://hacker-news.firebaseio.com/v0/item/" + itemId + ".json")).json();
            items.push(item);
        }
        console.log(items);
        this.setState({
            items: items
        })
    }
}
