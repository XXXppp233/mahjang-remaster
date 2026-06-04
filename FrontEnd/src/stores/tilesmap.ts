import { ref } from 'vue'
import { defineStore } from 'pinia'
// 编写 Python 代码时不知道麻将的名字该这么写，以后重写的时候可能会改为正确的名字。

interface TileMap {
  [key: string]: string
}
interface ActionsType {
  hu?: boolean | string // tile
  kong?: ActionChoice[] | ActionChoice | boolean | string // tile
  pong?: ActionChoice | boolean | string // tile
  chow?: ActionChoice[] | string[][]
}

interface ActionChoice {
  tile: string
  tiles?: string[]
  selec?: number[]
}

export const tilesmapStore = defineStore('tilemap', () => {
  const tilesmap = ref<TileMap>({
    Man1: 'Man1',
    Man2: 'Man2',
    Man3: 'Man3',
    Man4: 'Man4',
    Man5: 'Man5',
    Man6: 'Man6',
    Man7: 'Man7',
    Man8: 'Man8',
    Man9: 'Man9',
    Pin1: 'Pin1',
    Pin2: 'Pin2',
    Pin3: 'Pin3',
    Pin4: 'Pin4',
    Pin5: 'Pin5',
    Pin6: 'Pin6',
    Pin7: 'Pin7',
    Pin8: 'Pin8',
    Pin9: 'Pin9',
    Sou1: 'Sou1',
    Sou2: 'Sou2',
    Sou3: 'Sou3',
    Sou4: 'Sou4',
    Sou5: 'Sou5',
    Sou6: 'Sou6',
    Sou7: 'Sou7',
    Sou8: 'Sou8',
    Sou9: 'Sou9',
    Ton: 'Ton',
    Nan: 'Nan',
    Shaa: 'Shaa',
    Pei: 'Pei',
    Chun: 'Chun',
    Hatsu: 'Hatsu',
    Haku: 'Haku',
    Golden: 'Golden',
    Back: 'Back',
    Blank: 'Blank',
  })
  const fontsmap = ref<TileMap>({
    Pin1: '🀙',
    Pin2: '🀚',
    Pin3: '🀛',
    Pin4: '🀜',
    Pin5: '🀝',
    Pin6: '🀞',
    Pin7: '🀟',
    Pin8: '🀠',
    Pin9: '🀡',
    Sou1: '🀐',
    Sou2: '🀑',
    Sou3: '🀒',
    Sou4: '🀓',
    Sou5: '🀔',
    Sou6: '🀕',
    Sou7: '🀖',
    Sou8: '🀗',
    Sou9: '🀘',
    Man1: '🀇',
    Man2: '🀈',
    Man3: '🀉',
    Man4: '🀊',
    Man5: '🀋',
    Man6: '🀌',
    Man7: '🀍',
    Man8: '🀎',
    Man9: '🀏',
    Ton: '🀀',
    Nan: '🀁',
    Shaa: '🀂',
    Pei: '🀃',
    Haku: '🀆',
    Hatsu: '🀅',
    Chun: '🀄',
    Golden: '🃏',
    Back: '🀫',
    spring: '🀦',
    summer: '🀧',
    autumn: '🀨',
    winter: '🀩',
    plum: '🀢',
    orchid: '🀣',
    bamboo: '🀤',
    chrysanthemum: '🀥',
  })
  function getTileName(tile: string): string {
    if (tile) {
      return tilesmap.value[tile] || tile
    } else return ''
  }
  function getTilesName(tiles: string[]): string[] {
    if (tiles) {
      return tiles.map((tile) => getTileName(tile))
    } else return []
  }
  function getTileFont(tile: string | boolean): string {
    if (typeof tile === 'string') {
      return fontsmap.value[tile] || '�' // new tile may be ''
    } else return ''
  }
  function getTilesFont(tiles: string[]): string[] {
    if (tiles) {
      return tiles.map((tile) => getTileFont(tile))
    } else return []
  }
  function toChoices(
    action: ActionChoice[] | ActionChoice | boolean | string | string[] | string[][] | undefined,
  ): ActionChoice[] {
    if (!action) return []
    if (Array.isArray(action)) {
      if (action.length === 0) return []
      if (typeof action[0] === 'string') {
        return [{ tile: String(action[0]), tiles: action as unknown as string[] }]
      }
      if (Array.isArray(action[0])) {
        return (action as unknown as string[][]).map((tiles) => ({ tile: '', tiles }))
      }
      return action as ActionChoice[]
    }
    if (typeof action === 'string') return [{ tile: action }]
    if (typeof action === 'boolean') return []
    return [action]
  }

  function getActionsName(actions: ActionsType) {
    const actionsname = []
    if (actions) {
      if (actions.hu) {
        actionsname.push(`胡${getTileFont(actions.hu)}`)
      }
      if (actions.kong) {
        for (const choice of toChoices(actions.kong)) {
          actionsname.push(`杠${getTileFont(choice.tile)}`)
        }
      }
      if (actions.pong) {
        const choice = toChoices(actions.pong)[0]
        actionsname.push(`碰${getTileFont(choice?.tile ?? '')}`)
      }
      if (actions.chow) {
        for (const choice of toChoices(actions.chow as any)) {
          const tiles = choice.tiles ?? []
          actionsname.push(`吃${getTileFont(tiles[0] ?? '')}${getTileFont(tiles[1] ?? '')}`)
        }
      }
    } else {
      return []
    }
    return actionsname
  }
  function getActionData(actions: ActionsType) {
    const actionsdata = []
    if (actions.hu) {
      actionsdata.push(true)
    }
    if (actions.kong) {
      for (const choice of toChoices(actions.kong)) {
        actionsdata.push(choice.selec ?? true)
      }
    }
    if (actions.pong) {
      const choice = toChoices(actions.pong)[0]
      actionsdata.push(choice?.selec ?? true)
    }
    if (actions.chow) {
      for (const choice of toChoices(actions.chow as any)) {
        actionsdata.push(choice.selec ?? choice.tiles ?? [])
      }
    }
    console.log('getActionData', actions, actionsdata)
    return actionsdata
  }

  return {
    tilesmap,
    getTileName,
    getTilesName,
    fontsmap,
    getTileFont,
    getTilesFont,
    getActionsName,
    getActionData,
  }
})
